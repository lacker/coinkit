package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"

	"coinkit/auth"
	"coinkit/network"
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
	keyPair *auth.KeyPair
	peers []*network.Peer
	info map[string]*PeerInfo
}

func NewServer(c *Config) *Server {
	var peers []*network.Peer
	for _, p := range c.PeerPorts {
		peers = append(peers, network.NewPeer(p))
	}
	
	return &Server{
		port: c.Port,
		keyPair: c.KeyPair,
		peers: peers,
		info: make(map[string]*PeerInfo),
	}
}

// Handles an incoming connection
func (s *Server) handleConnection(conn net.Conn) {
	log.Printf("handling a connection")
	for {
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			conn.Close()
			break
		}
		// Chop the newline
		serialized := data[:len(data)-1]
		sm, err := auth.NewSignedMessageFromSerialized(serialized)
		if err != nil {
			// The signature isn't valid.
			// Maybe the message got chopped off? Maybe they are bad guys?
			// Assume good intentions and close the connection.
			log.Printf("got bad data: [%s]", data)
			conn.Close()
			break
		}

		// Get the info for this peer
		info, ok := s.info[sm.Signer()]
		if !ok {
			info = NewPeerInfo(sm.Signer())
			s.info[info.publicKey] = info
		}
		
		switch m := sm.Message().(type) {
		case *network.UptimeMessage:
			if m.Uptime > info.uptime {
				// As it should be
				info.uptime = m.Uptime
			} else if m.Uptime == info.uptime {
				log.Printf("duplicate message for uptime %d from %s",
					info.uptime, info.publicKey)
			} else {
				log.Printf("node %s appears to have restarted", info.publicKey)
			}
		default:
			log.Printf("could not handle message: %s", network.EncodeMessage(m))
		}
				
		fmt.Fprintf(conn, "ok\n")
	}
}

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

func (s *Server) ServeForever() {
	go s.listen()

	uptime := 0
	for {
		time.Sleep(time.Second)
		log.Printf("uptime is %d", uptime)
		for _, peer := range s.peers {
			message := &network.UptimeMessage{Uptime: uptime}
			sm := auth.NewSignedMessage(s.keyPair, message)
			peer.Send(sm.Serialize())
		}
		uptime++
	}
}
