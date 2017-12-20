package main

import (
	"log"
	"os"
	"strconv"

	"coinkit/server"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <i> where i is in [0, 1, 2, ..., NumPeers - 1]")
	}
	arg, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	config := server.NewLocalConfig(arg)
	
	s := server.NewServer(config)
	s.ServeForever()
}
