package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"coinkit/currency"
	"coinkit/util"
)

type Server struct {
	port    int
	keyPair *util.KeyPair
	peers   []*Client
	node    *Node

	// Whenever there is a new batch of outgoing messages, it is serialized
	// into a list of lines and sent to the outgoing channel
	outgoing chan []string

	// Messages we are going to handle. These do not require a response
	messages chan *util.SignedMessage

	// Requests we are going to handle. These require a response
	requests chan *Request

	listener net.Listener

	// We close the currentBlock channel whenever the current block is complete
	currentBlock chan bool

	// We close the quit channel and set shutdown to true
	// when the server is shutting down
	shutdown bool
	quit     chan bool

	// How often we send out a broadcast of redundant data
	BroadcastInterval time.Duration
}

func NewServer(config *ServerConfig) *Server {
	peers := []*Client{}
	for _, address := range config.Network.Nodes {
		peers = append(peers, NewClient(address))
	}
	qs := config.Network.QuorumSlice()

	// At the start, all money is in the "mint" account
	node := NewNode(config.KeyPair.PublicKey(), qs)
	mint := util.NewKeyPairFromSecretPhrase("mint")
	node.queue.SetBalance(mint.PublicKey(), currency.TotalMoney)

	return &Server{
		port:              config.Port,
		keyPair:           config.KeyPair,
		peers:             peers,
		node:              node,
		outgoing:          make(chan []string, 10),
		messages:          make(chan *util.SignedMessage),
		requests:          make(chan *Request),
		listener:          nil,
		shutdown:          false,
		quit:              make(chan bool),
		currentBlock:      make(chan bool),
		BroadcastInterval: time.Second,
	}
}

func (s *Server) Logf(format string, a ...interface{}) {
	util.Logf("SE", s.keyPair.PublicKey(), format, a...)
}

// Handles an incoming connection.
// This is likely to include many messages, all separated by endlines.
func (s *Server) handleConnection(conn net.Conn) {
	for {
		sm, err := util.ReadSignedMessage(conn)
		if err != nil {
			if !s.shutdown && err != io.EOF {
				log.Printf("connection error: %v", err)
			}
			conn.Close()
			break
		}

		if sm == nil {
			continue
		}

		// Send this message to the processing goroutine
		response := make(chan *util.SignedMessage)
		request := &Request{
			Message:  sm,
			Response: response,
		}

		// Send our request to the processing goroutine, wait for the response,
		// and return it down the connection
		s.requests <- request
		timer := time.NewTimer(time.Second)
		select {
		case m := <-response:
			util.WriteSignedMessage(conn, m)
		case <-s.quit:
			conn.Close()
			break
		case <-timer.C:
			log.Fatalf("the processing goroutine got overloaded")
		}
	}
}

// Flushes the outgoing queue and returns the last value if there is any.
// Returns [], false if there is none
// Does not wait
func (s *Server) getOutgoing() ([]string, bool) {
	lines := []string{}
	ok := false
	for {
		select {
		case lines = <-s.outgoing:
			ok = true
		default:
			return lines, ok
		}
	}
}

// unsafeUpdateOutgoing gets the outgoing messages from our node and uses
// the outgoing channel to broadcast them.
// Since it deals with the node directly, it should only be called from the
// message-processing goroutine.
func (s *Server) unsafeUpdateOutgoing() {
	// First encode the outgoing messages into lines
	out := s.node.OutgoingMessages()

	lines := []string{}
	for _, m := range out {
		sm := util.NewSignedMessage(s.keyPair, m)
		lines = append(lines, util.SignedMessageToLine(sm))
	}

	// Clear the outgoing queue
	s.getOutgoing()

	// Send our lines to the now-probably-empty queue
	s.outgoing <- lines
}

// unsafeHandleMessage handles a message by interacting with the node directly.
// It should be only be called from the message-processing goroutine.
func (s *Server) unsafeHandleMessage(m *util.SignedMessage) *util.SignedMessage {
	prevSlot := s.node.Slot()
	message := s.node.Handle(m.Signer(), m.Message())
	postSlot := s.node.Slot()
	s.unsafeUpdateOutgoing()

	if postSlot != prevSlot {
		close(s.currentBlock)
		s.currentBlock = make(chan bool)
	}

	// Return the appropriate message
	if message == nil {
		return nil
	}
	sm := util.NewSignedMessage(s.keyPair, message)
	return sm
}

// processMessagesForever should be run in its own goroutine. This is the only
// goroutine that is allowed to access the node, because node is not threadsafe.
// The 'unsafe' methods should only be called from within here.
func (s *Server) processMessagesForever() {
	for {

		select {

		case request := <-s.requests:
			if request.Message != nil {
				response := s.unsafeHandleMessage(request.Message)
				if request.Response != nil {
					request.Response <- response
				}
			}

		case message := <-s.messages:
			if message != nil {
				s.unsafeHandleMessage(message)
			}

		case <-s.quit:
			break
		}
	}
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if s.shutdown {
			break
		}
		if err != nil {
			log.Print("incoming connection error: ", err)
		}
		go s.handleConnection(conn)
	}
}

// Must be called before listen()
// Will retry up to 5 seconds
func (s *Server) acquirePort() {
	s.Logf("listening on port %d", s.port)
	for i := 0; i < 100; i++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
		if err == nil {
			s.listener = ln
			return
		}
		time.Sleep(time.Millisecond * time.Duration(50))
	}
	log.Fatalf("could not acquire port %d", s.port)
}

func (s *Server) broadcastLines(lines []string) {
	for _, line := range lines {
		for _, peer := range s.peers {
			peer.Send(&Request{
				Line:     line,
				Response: s.messages,
				Timeout:  5 * time.Second,
			})
		}
	}
}

func scontains(list []string, s string) bool {
	for _, str := range list {
		if str == s {
			return true
		}
	}
	return false
}

// broadcastIntermittently() sends outgoing messages every so often. it
// should be run as a goroutine.
func (s *Server) broadcastIntermittently() {
	lastLines := []string{}

	for {
		timer := time.NewTimer(s.BroadcastInterval)
		select {

		case <-s.quit:
			break

		case lines := <-s.outgoing:

			// See if there are even newer lines
			newerLines, ok := s.getOutgoing()
			if ok {
				lines = newerLines
			}

			// When we receive a new outgoing, we only need to send out the
			// lines that have changed since last time.
			changedLines := []string{}
			for _, line := range lines {
				if !scontains(lastLines, line) {
					changedLines = append(changedLines, line)
				}
			}

			lastLines = lines
			s.broadcastLines(changedLines)

		case <-timer.C:
			// When we hit the timer, we rebroadcast the whole outbox.
			// This is a backstop against miscellaneous problems. If the
			// network is functioning perfectly, this isn't necessary.
			s.broadcastLines(lastLines)
		}
	}
}

func (s *Server) LocalhostAddress() *Address {
	return &Address{
		Host: "localhost",
		Port: s.port,
	}
}

// ServeForever spawns off all the goroutines and never returns.
// Stop() might not work when you run the server this way, because stopping
// during startup does not work well
func (s *Server) ServeForever() {
	s.acquirePort()

	go s.processMessagesForever()
	go s.listen()
	s.broadcastIntermittently()
}

// ServeInBackground spawns goroutines to run the server.
// It returns once it has successfully bound to its port.
// Stop() should work if it is called after ServeInBackground returns.
func (s *Server) ServeInBackground() {
	s.acquirePort()
	go s.processMessagesForever()
	go s.listen()
	go s.broadcastIntermittently()
}

func (s *Server) Stop() {
	s.shutdown = true
	close(s.quit)

	if s.listener != nil {
		log.Printf("closing listener on port %d", s.port)
		s.listener.Close()
	}
}
