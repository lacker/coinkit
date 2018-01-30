package network

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"coinkit/consensus"
	"coinkit/currency"
	"coinkit/util"
)

type Server struct {
	port     int
	keyPair  *util.KeyPair
	peers    []*Client
	node     *Node
	outgoing []util.Message

	// Messages we are going to handle. These do not require a response
	messages chan *util.SignedMessage

	// Requests we are going to handle. These require a response
	requests chan *Request
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
		port:     c.Port,
		keyPair:  c.KeyPair,
		peers:    peers,
		node:     node,
		outgoing: node.OutgoingMessages(),
		messages: make(chan *util.SignedMessage),
		requests: make(chan *Request),
	}
}

// Handles an incoming connection.
// This is likely to include many messages, all separated by endlines.
func (s *Server) handleConnection(conn net.Conn) {
	for {
		sm, err := util.ReadSignedMessage(conn)
		if err != nil {
			// Assume good intentions and close the connection.
			log.Printf("connection error: %v", err)
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
		m := <-response

		util.WriteSignedMessage(conn, m)
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

		}
	}
}

// listen() runs a server that spawns a goroutine for each client that connects
func (s *Server) listen(errChan chan error) {
	log.Printf("listening on port %d", s.port)
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
	if err != nil {
		log.Print(err)
		errChan <- err
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print("incoming connection error: ", err)
		}
		go s.handleConnection(conn)
	}
}

// ServeForever spawns off all the goroutines
func (s *Server) ServeForever() error {
	go s.handleMessagesForever()

	listenErrChan := make(chan error)
	go s.listen(listenErrChan)

	listenErr := <-listenErrChan
	if (listenErr != nil) {
		return listenErr
	}

	for {
		// TODO: go faster if we have new info
		time.Sleep(time.Second * time.Duration(1+rand.Float64()))

		// Broadcast to all peers
		// Don't use s.outgoing directly in case the listen() goroutine
		// modifies it while we iterate on it
		messages := s.outgoing
		for _, m := range messages {
			sm := util.NewSignedMessage(s.keyPair, m)
			for _, peer := range s.peers {
				peer.Send(&Request{
					Message:  sm,
					Response: s.messages,
				})
			}
		}
	}
}
