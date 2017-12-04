package network

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

type Server struct {
	port int
	peers []*Peer
}

func NewServer(port int, peers []*Peer) *Server {
	return &Server{port: port, peers: peers}
}

// Handles an incoming connection
func (s *Server) handleConnection(conn net.Conn) {
	log.Printf("handling a connection")
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			conn.Close()
			break
		}
		log.Printf("got message: %s", message)
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
			peer.Send(fmt.Sprintf("broadcasting uptime %d", uptime))
		}
		uptime++
	}
}
