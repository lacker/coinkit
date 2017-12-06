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

type Server struct {
	port int
	keyPair *auth.KeyPair
	peers []*network.Peer
}

func NewServer(port int, kp *auth.KeyPair, peers []*network.Peer) *Server {
	return &Server{port: port, keyPair: kp, peers: peers}
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
		log.Printf("got message: %s", network.EncodeMessage(sm.Message()))
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
