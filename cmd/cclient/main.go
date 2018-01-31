package main 

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	
	"coinkit/currency"
	"coinkit/network"
	"coinkit/util"
)

func newClient() *network.Client {
	config, _ := network.NewLocalNetwork()
	address := config.RandomAddress()
	c := network.NewClient(address)
	log.Printf("connecting to %s", address.String())
	return c
}

// Fetches, displays, and returns the status for a user.
func status(user string) *currency.Account {
	client := newClient()
	account := client.GetAccount(user)

	log.Printf("account data for %s:\n%s", user, spew.Sdump(account))
	return account
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
	kp := util.NewKeyPairFromSecretPhrase(phrase)
	log.Printf("hello. your name is %s", kp.PublicKey())
	return kp
}

func send(recipient string, amountStr string) {
	amountInt, err := strconv.Atoi(amountStr)
	if err != nil {
		log.Fatalf("could not convert %s to a number", amountStr)
	}
	amount := uint64(amountInt)
	kp := login()
	user := kp.PublicKey()
	client := newClient()
	account := client.GetAccount(user)

	log.Printf("account data for %s:\n%s", user, spew.Sdump(account))
	
	if account.Balance < amount {
		log.Fatalf("cannot send %d when our account only has %d",
			amount, account.Balance)
	}

	transaction := &currency.Transaction{
		From: user,
		Sequence: account.Sequence + 1,
		To: recipient,
		Amount: amount,
		Fee: 0,
	}

	// Send our transaction to the network
	st := transaction.SignWith(kp)
	tm := currency.NewTransactionMessage(st)
	sm := util.NewSignedMessage(kp, tm)
	client.SendMessage(sm)
	log.Printf("sending %d to %s", amount, recipient)
	
	// Wait for our transaction to clear
	for {
		a := status(user)
		if a.Sequence >= transaction.Sequence {
			break
		}
		time.Sleep(time.Second)
	}
	log.Printf("transaction %d cleared", transaction.Sequence)
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
