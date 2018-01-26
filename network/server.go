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
	port int
	keyPair *util.KeyPair
	peers []*Client
	node *Node
	outgoing []util.Message

	// Messages get validated before entering the inbox
	inbox chan *util.SignedMessage
}

func NewServer(c *Config) *Server {
	inbox := make(chan *util.SignedMessage)
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
		inbox: inbox,
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

		// Send this message to the processing goroutine
		s.inbox <- sm
		
		util.WriteNilMessageTo(conn)
	}
}

// handleMessage should only be called by a single goroutine, because the
// node objects aren't threadsafe.
// Caller should be validating the signature
func (s *Server) handleMessage(sm *util.SignedMessage) {
	// TODO: send back any response messages
	s.node.Handle(sm.Signer(), sm.Message())
	s.outgoing = s.node.OutgoingMessages()
}

func (s *Server) handleMessagesForever() {
	for {
		m := <-s.inbox
		s.handleMessage(m)
	}
}

// listen() runs a server that spawns a goroutine for each client that connects
func (s *Server) listen() {
	log.Printf("listening on port %d", s.port)
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
	if err != nil {
		log.Fatal(err)
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
func (s *Server) ServeForever() {
	go s.handleMessagesForever()
	go s.listen()

	for {
		// TODO: go faster if we have new info
		time.Sleep(time.Second * time.Duration(1 + rand.Float64()))

		// Broadcast to all peers
		// Don't use s.outgoing directly in case the listen() goroutine
		// modifies it while we iterate on it
		messages := s.outgoing
		for _, m := range messages {
			sm := util.NewSignedMessage(s.keyPair, m)
			for _, peer := range s.peers {
				peer.Send(sm)
			}
		}
	}
}
