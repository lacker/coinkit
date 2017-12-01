package main

import "bufio"
import "fmt"
import "log"
import "net"
import "os"
import "strconv"
import "time"

import "coinkit/network"

const (
	BasePort = 9000
	Nodes     = 2
)

// Handles an incoming connection
func handleConnection(conn net.Conn) {
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

func listen(port int) {
	log.Printf("listening on port %d", port)
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print("incoming connection error: ", err)
		}
		go handleConnection(conn)
	}
}

func main() {
	// Usage: go run main.go <i> where i is in [0, 1, 2, ..., Nodes - 1]
	if len(os.Args) < 2 {
		log.Fatal("Use an argument with a numerical id.")
	}
	id, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if id < 0 || id >= Nodes {
		log.Fatalf("invalid id: %d", id)
	}

	port := BasePort + id
	log.Printf("server %d starting up on port %d", id, port)

	// Make some peers
	var peers []*network.Peer
	for p := BasePort; p < BasePort+Nodes; p++ {
		if p == port {
			continue
		}
		peer := network.NewPeer(p)
		peers = append(peers, peer)
	}

	go listen(port)

	uptime := 0
	for {
		time.Sleep(time.Second)
		log.Printf("uptime is %d", uptime)
		for _, peer := range peers {
			peer.Send(fmt.Sprintf("node %d has uptime %d", id, uptime))
		}
		uptime++
	}
}
