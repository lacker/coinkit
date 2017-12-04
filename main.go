package main

import (
	"log"
	"os"
	"strconv"
	
	"coinkit/network"
)

const (
	BasePort = 9000
	NumPeers = 2
)

func main() {
	// Usage: go run main.go <i> where i is in [0, 1, 2, ..., NumPeers - 1]
	if len(os.Args) < 2 {
		log.Fatal("Use an argument with a numerical id.")
	}
	id, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if id < 0 || id >= NumPeers {
		log.Fatalf("invalid id: %d", id)
	}

	port := BasePort + id


	// Make some peers
	var peers []*network.Peer
	for p := BasePort; p < BasePort+NumPeers; p++ {
		if p == port {
			continue
		}
		peer := network.NewPeer(p)
		peers = append(peers, peer)
	}

	server := network.NewServer(port, peers)
	server.ServeForever()
}
