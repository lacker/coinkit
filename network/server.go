package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"coinkit/currency"
	"coinkit/util"
)

type Server struct {
	port     int
	keyPair  *util.KeyPair
	peers    []*Client
	node     *Node

	// Whenever there is a new batch of outgoing messages, it is serialized
	// into a list of lines and sent to the outgoing channel
	outgoing chan []string

	// Messages we are going to handle. These do not require a response
	messages chan *util.SignedMessage

	// Requests we are going to handle. These require a response
	requests chan *Request

	listener net.Listener

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
		port:     config.Port,
		keyPair:  config.KeyPair,
		peers:    peers,
		node:     node,
		outgoing: make(chan []string, 10),
		messages: make(chan *util.SignedMessage),
		requests: make(chan *Request),
		listener: nil,
		shutdown: false,
		quit: make(chan bool),
		BroadcastInterval: time.Second,
	}
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
		timer := time.NewTimer(time.Second * 5)
		select {
		case m := <-response:
			util.WriteSignedMessage(conn, m)
		case <-s.quit:
			conn.Close()
			break
		case <-timer.C:
			log.Fatalf("we failed to respond to a message within 5 seconds")
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

func (s *Server) updateOutgoing() {
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

func (s *Server) handleMessage(m *util.SignedMessage) *util.SignedMessage {
	message := s.node.Handle(m.Signer(), m.Message())
	s.updateOutgoing()

	// Return the appropriate message
	if message == nil {
		return nil
	}
	sm := util.NewSignedMessage(s.keyPair, message)
	return sm
}

func (s *Server) handleMessagesForever() {
	for {

		select {

		case request := <-s.requests:
			if request.Message != nil {
				response := s.handleMessage(request.Message)
				if request.Response != nil {
					request.Response <- response
				}
			}

		case message := <-s.messages:
			if message != nil {
				s.handleMessage(message)
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
	log.Printf("listening on port %d", s.port)
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
				Line: line,
				Response: s.messages,
			})
		}
	}
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

			if strings.Join(lastLines, ",") == strings.Join(lines, ",") {
				// It's just the same thing, no need for an instant send
				continue
			}

			lastLines = lines
			s.broadcastLines(lines)

		case <-timer.C:
			// Re-send stuff
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

	go s.handleMessagesForever()
	go s.listen()
	s.broadcastIntermittently()
}

// ServeInBackground spawns goroutines to run the server.
// It returns once it has successfully bound to its port.
// Stop() should work if it is called after ServeInBackground returns.
func (s *Server) ServeInBackground() {
	s.acquirePort()
	go s.handleMessagesForever()
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
