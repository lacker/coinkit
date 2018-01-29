package main 

import (
	"bufio"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	
	"coinkit/currency"
	"coinkit/network"
	"coinkit/util"
)

// getAccount fetches account data from a server and sends it, once, to the
// provided channel.
func getAccount(user string, output chan *currency.Account) {
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
	sm = <-response
	m := sm.Message()
	am, ok := m.(*currency.AccountMessage)
	if !ok {
		log.Fatal("received non-account message: %+v", message)
	}
	account := am.State[user]
	output <- account
}

// Fetches and displays the status for a user.
func status(user string) {
	ac := make(chan *currency.Account)
	go getAccount(user, ac)
	account := <-ac

	log.Printf("account data for %s:\n%s", user, spew.Sdump(account))
}

// Asks for a login then displays the status
func ourStatus() {
	kp := login()
	status(kp.PublicKey())
}

// Ask the user for a passphrase to log in.
func login() *util.KeyPair {
	log.Printf("please enter your passphrase:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
    phrase := scanner.Text()
	log.Printf("read phrase: [%s]", phrase)
	kp := util.NewKeyPairFromSecretPhrase(phrase)
	log.Printf("hello. your name is %s", kp.PublicKey())
	return kp
}

func send(recipient string, amount string) {
	kp := login()
	log.Printf("kp: %+v", kp)
	// TODO: fetch our own data so that we know what sequence number to use
}

// cclient runs a client that connects to the coinkit network.
func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: cclient {send,status} ...")
	}
	op := os.Args[1]
	rest := os.Args[2:]
	switch op {
	case "status":
		if len(rest) > 1 {
			log.Fatal("Usage: cclient status [publickey]")
		}
		if len(rest) == 0 {
			ourStatus()
		} else {
			status(rest[0])
		}
	case "send":
		if len(rest) != 2 {
			log.Fatal("Usage: cclient send <user> <amount>")
		}
		send(rest[0], rest[1])
	default:
		log.Fatalf("unrecognized operation: %s", op)
	}
}
