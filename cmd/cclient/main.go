package main

import (
	"bufio"
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/crypto/sha3"

	"coinkit/currency"
	"coinkit/data"
	"coinkit/network"
	"coinkit/util"
	"fmt"
	"net/http"
	"strings"
)

func newConnection() network.Connection {
	config, _ := network.NewLocalNetwork()
	address := config.RandomAddress()
	c := NewRedialConnection(address, nil)
	log.Printf("connecting to %s", address.String())
	return c
}

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
	status(kp.PublicKey().String())
}

// Ask the user for a passphrase to log in.
func login() *util.KeyPair {
	log.Printf("please enter your passphrase:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	phrase := scanner.Text()
	kp := util.NewKeyPairFromSecretPhrase(phrase)
	log.Printf("hello. your name is %s", kp.PublicKey().String())
	return kp
}

func send(recipient string, amountStr string) {
	amountInt, err := strconv.Atoi(amountStr)
	if err != nil {
		log.Fatalf("could not convert %s to a number", amountStr)
	}
	if _, err := util.ReadPublicKey(recipient); err != nil {
		log.Fatalf("invalid address: %s", recipient)
	}
	amount := uint64(amountInt)
	kp := login()
	user := kp.PublicKey().String()
	client := newClient()
	account := client.GetAccount(user)

	log.Printf("account data for %s:\n%s", user, spew.Sdump(account))

	if account.Balance < amount {
		log.Fatalf("cannot send %d when our account only has %d",
			amount, account.Balance)
	}

	seq := account.Sequence + 1
	transaction := &currency.Transaction{
		From:     user,
		Sequence: seq,
		To:       recipient,
		Amount:   amount,
		Fee:      0,
	}

	// Send our transaction to the network
	st := transaction.SignWith(kp)
	tm := currency.NewTransactionMessage(st)
	sm := util.NewSignedMessage(kp, tm)
	client.SendMessage(sm)
	log.Printf("sending %d to %s", amount, recipient)

	// Wait for our transaction to clear
	client.WaitToClear(user, seq)
	log.Printf("transaction %d cleared", transaction.Sequence)
}

func handler(w http.ResponseWriter, r *http.Request) {
	pass := strings.TrimLeft(r.URL.Path, "/")
	kp := util.NewKeyPairFromSecretPhrase(pass)
	s := status(kp.PublicKey().String())
	if s != nil {
		fmt.Fprintf(w, "{ \"sequence\": %d, \"balance\": %d }",
			s.Sequence, s.Balance)
	} else {
		fmt.Fprintf(w, "{}")
	}
}

func proxy() {
	log.Printf("Running client proxy on port 9090")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":9090", nil)
}

func upload(filename string) {
	// Construct a message from the file
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	h := sha3.New512()
	h.Write(bytes)
	checksum := h.Sum(nil)
	key := hex.EncodeToString(checksum[:8])
	log.Printf("uploading file as: %s", key)
	dmap := make(map[string]string)
	dmap[key] = string(bytes)
	message := &data.DataMessage{
		Data: dmap,
	}

	kp := util.NewKeyPair()
	client := newClient()
	sm := util.NewSignedMessage(kp, message)
	client.SendMessage(sm)
}

// cclient runs a client that connects to the coinkit network.
func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: cclient {send,status,proxy} ...")
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
	case "proxy":
		proxy()
	case "upload":
		if len(rest) != 1 {
			log.Fatal("Usage: cclient upload <filename>")
		}
		upload(rest[0])
	default:
		log.Fatalf("unrecognized operation: %s", op)
	}
}
