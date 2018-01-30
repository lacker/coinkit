package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"coinkit/consensus"
	"coinkit/currency"
	"coinkit/util"
)

type Server struct {
	port int
	keyPair *util.KeyPair
	peers []*Client
	node *Node
	outgoing []util.Message

	// Messages we are going to handle. These do not require a response
	messages chan *util.SignedMessage
	
	// Requests we are going to handle. These require a response
	requests chan *Request

	listener net.Listener

	// We close the quit channel and set shutdown to true
	// when the server is shutting down
	shutdown bool
	quit chan bool
}

func NewServer(c *Config) *Server {
	var peers []*Client
	log.Printf("config has peers: %v", c.PeerPorts)
	for _, p := range c.PeerPorts {
		peers = append(peers, NewClient(p))
	}

	qs := consensus.MakeQuorumSlice(c.Members, c.Threshold)
	
	// At the start, all money is in the "mint" account
	node := NewNode(c.KeyPair.PublicKey(), qs)
	mint := util.NewKeyPairFromSecretPhrase("mint")
	log.Printf("establishing a mint: %s", mint.PublicKey())
	node.queue.SetBalance(mint.PublicKey(), currency.TotalMoney)
	
	return &Server{
		port: c.Port,
		keyPair: c.KeyPair,
		peers: peers,
		node: node,
		outgoing: node.OutgoingMessages(),
		messages: make(chan *util.SignedMessage),
		requests: make(chan *Request),
		listener: nil,
		shutdown: false,
		quit: make(chan bool),
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
			Message: sm,
			Response: response,
		}

		// Send our request to the processing goroutine, wait for the response,
		// and return it down the connection
		s.requests <- request
		select {
		case m := <-response:
			util.WriteSignedMessage(conn, m)
		case <-s.quit:
			conn.Close()
			break
		}
	}
}

func (s *Server) handleMessage(m *util.SignedMessage) *util.SignedMessage {
	message := s.node.Handle(m.Signer(), m.Message())
	s.outgoing = s.node.OutgoingMessages()
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

// listen() runs a server that spawns a goroutine for each client that connects
func (s *Server) listen() {
	for {
		log.Printf("accepting on port %d", s.port)		
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
func (s *Server) acquirePort() {
	log.Printf("listening on port %d", s.port)
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
	if err != nil {
		log.Fatal(err)
	}
	s.listener = ln	
}

// broadcastIntermittently() sends outgoing messages every so often. it
// should be run as a goroutine.
func (s *Server) broadcastIntermittently() {
	for {
		// TODO: go faster if we have new info
		time.Sleep(time.Second)
		if s.shutdown {
			break
		}
		
		// Broadcast to all peers
		// Don't use s.outgoing directly in case the listen() goroutine
		// modifies it while we iterate on it
		messages := s.outgoing
		for _, m := range messages {
			sm := util.NewSignedMessage(s.keyPair, m)
			for _, peer := range s.peers {
				peer.Send(&Request{
					Message: sm,
					Response: s.messages,
				})
			}
		}
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
