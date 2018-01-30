package main

import (
	"log"
	"os"
	"strconv"

	"coinkit/network"
)

// cserver runs a coinkit server.

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: cserver <i> where i is in [0, 1, 2, ..., NumPeers - 1]")
	}
	arg, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	config := network.NewLocalConfig(arg)
	
	s := network.NewServer(config)
	s.ServeForever()
}
