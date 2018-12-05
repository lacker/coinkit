package main

import (
	"github.com/lacker/coinkit/hosting"
)

// chost runs a p2p hosting server.
// probably in the future we want to combine this with cserver. this is useful for testing
// just the p2p hosting protocols though, for now.

func main() {
	server := hosting.NewServer()
	server.Serve()
}
