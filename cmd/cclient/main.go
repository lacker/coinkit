package main 

import (
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	
	"coinkit/currency"
	"coinkit/network"
	"coinkit/util"
)

// Fetches and displays the status for a user.
func status(user string) {
	// Since this is public data we'll use a throwaway key and stay anonymous
	kp := util.NewKeyPair()

	message := currency.NewInquiryMessage(user)
	sm := util.NewSignedMessage(kp, message)
	client := network.NewClient(network.RandomLocalServer())
	response := make(chan *util.SignedMessage)
	request := &network.Request{
		Message: sm,
		Response: response,
	}

	// Wait on a response.
	// This hangs on network failure
	client.Send(request)
	m := <-response
	
	log.Printf("response: %s", spew.Sdump(m))
}

// cclient runs a client that connects to the coinkit network.
func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: cclient status <publickey>")
	}
	op := os.Args[1]
	rest := os.Args[2:]
	switch op {
	case "status":
		status(rest[0])
	default:
		log.Fatalf("unrecognized operation: %s", op)
	}
}
