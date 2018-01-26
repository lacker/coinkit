package main 

import (
	"log"

	"coinkit/currency"
	"coinkit/network"
)

// Fetches and displays the status for a user.
func status(user string) {
	message := currency.NewInquiryMessage(user)
	peer := network.NewPeer(network.RandomLocalServer())
		
	// TODO: get a response
	log.Fatalf("%+v %+v", message, peer)
}

// cclient runs a client that connects to the coinkit network.
func main() {
	log.Printf("TODO: implement something here")
}
