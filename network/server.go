package network

import (
	"bufio"
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
	inbox chan *util.SignedMessage
}

func NewServer(c *Config) *Server {
	var peers []*Peer
	log.Printf("config has peers: %v", c.PeerPorts)
	for _, p := range c.PeerPorts {
		peers = append(peers, NewPeer(p))
	}

	qs := consensus.MakeQuorumSlice(c.Members, c.Threshold)
	
	// At the start, all money is in the "mint" account
	node := NewNode(c.KeyPair.PublicKey(), qs)
	mint := util.NewKeyPairFromSecretPhrase("mint")
	node.queue.SetBalance(mint.PublicKey(), currency.TotalMoney)
	
	return &Server{
		port: c.Port,
		keyPair: c.KeyPair,
		peers: peers,
		info: make(map[string]*PeerInfo),
		node: node,
		outgoing: node.OutgoingMessages(),
		inbox: make(chan *util.SignedMessage),
	}
}

// Handles an incoming connection.
// This is likely to include many messages, all separated by endlines.
func (s *Server) handleConnection(conn net.Conn) {
	for {
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			conn.Close()
			break
		}

		// Chop the newline
		serialized := data[:len(data)-1]
		sm, err := util.NewSignedMessageFromSerialized(serialized)
		if err != nil {
			// The signature isn't valid.
			// Maybe the message got chopped off? Maybe they are bad guys?
			// Assume good intentions and close the connection.
			log.Printf("got %d bytes of bad data: [%s]", len(serialized), serialized)
			log.Printf("error: %v", err)
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
		
		fmt.Fprintf(conn, "ok\n")
	}
}

// handleMessage should only be called by a single goroutine, because the
// node objects aren't threadsafe.
// Caller should be validating the signature
func (s *Server) handleMessage(sm *util.SignedMessage) {
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
	line := sm.Serialize()
	// log.Printf("sending %d bytes of data: [%s]", len(line), line)
	for _, peer := range s.peers {
		peer.Send(line)
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
