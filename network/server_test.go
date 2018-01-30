package network

import (
	"log"
	"testing"
	"time"

	"coinkit/currency"
	"coinkit/util"
)

func makeServers() []*Server {
	answer := []*Server{}
	for i := 0; i <= 3; i++ {
		config := NewLocalConfig(i)
		server := NewServer(config)
		server.ServeInBackground()
		answer = append(answer, server)
	}
	return answer
}

func stopServers(servers []*Server) {
	for _, server := range servers {
		server.Stop()
	}
}

func TestStartStop(t *testing.T) {
	servers := makeServers()
	stopServers(servers)
	moreServers := makeServers()
	stopServers(moreServers)
}

func TestSendingMoney(t *testing.T) {
	servers := makeServers()
	mint := util.NewKeyPairFromSecretPhrase("mint")
	bob := util.NewKeyPairFromSecretPhrase("bob")
	transaction := &currency.Transaction{
		From: mint.PublicKey(),
		Sequence: 1,
		To: bob.PublicKey(),
		Amount: 100,
		Fee: 1,
	}
	st := transaction.SignWith(mint)
	tm := currency.NewTransactionMessage(st)
	sm := util.NewSignedMessage(mint, tm)
	client := NewClient(9000)
	client.SendMessage(sm)
	log.Printf("sent a message")
	
	failures := 0
	for {
		account := client.GetAccount(bob.PublicKey())
		log.Printf("got account: %+v", account)
		
		if account != nil && account.Balance > 0 {
			break
		}
		failures++

		log.Printf("%d failures", failures)
		if failures >= 10 {
			t.Fatalf("too much failure")
		}
		
		time.Sleep(time.Second)
	}

	stopServers(servers)
}
