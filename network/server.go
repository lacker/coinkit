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

type PeerInfo struct {
	publicKey string
	uptime int
}

func NewPeerInfo(publicKey string) *PeerInfo {
	return &PeerInfo{
		publicKey: publicKey,
		uptime: 0,
	}
}

type Server struct {
	port int
	keyPair *util.KeyPair
	peers []*Peer
	info map[string]*PeerInfo
	node *Node
	outgoing []util.Message

	// Messages get validated before entering the inbox
	inbox chan *util.SignedMessage
}

func NewServer(c *Config) *Server {
	inbox := make(chan *util.SignedMessage)
	var peers []*Peer
	log.Printf("config has peers: %v", c.PeerPorts)
	for _, p := range c.PeerPorts {
		peers = append(peers, NewPeer(p))
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
		info: make(map[string]*PeerInfo),
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

		// Get the info for this peer
		info, ok := s.info[sm.Signer()]
		if !ok {
			info = NewPeerInfo(sm.Signer())
			s.info[info.publicKey] = info
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

func (s *Server) broadcast(m util.Message) {
	sm := util.NewSignedMessage(s.keyPair, m)
	for _, peer := range s.peers {
		peer.Send(sm)
	}
}

// ServeForever spawns off all the goroutines
func (s *Server) ServeForever() {
	go s.handleMessagesForever()
	go s.listen()

	for {
		time.Sleep(time.Second * time.Duration(1 + rand.Float64()))
		// Don't use s.outgoing directly in case the listen() goroutine
		// modifies it while we iterate on it
		messages := s.outgoing
		for _, message := range messages {
			s.broadcast(message)
		}
	}
}
