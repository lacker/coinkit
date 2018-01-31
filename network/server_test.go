package network

import (
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"coinkit/currency"
	"coinkit/util"
)

// FakeMessage implements util.Message but does not get registered
type FakeMessage struct {
	Number int
}

func (m *FakeMessage) Slot() int {
	return 0
}

func (m *FakeMessage) MessageType() string {
	return "Fake"
}

func makeServers() []*Server {
	_, configs := NewUnitTestNetwork()
	answer := []*Server{}
	for _, config := range configs {
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
	client := NewClient(servers[0].LocalhostAddress())
	client.SendMessage(sm)
	
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

	go stopServers(servers)
}

func TestServerOkayWithFakeWellFormattedMessage(t *testing.T) {
	_, configs := NewUnitTestNetwork()
	s := NewServer(configs[0])

	m := &FakeMessage{Number: 4}
	kp := util.NewKeyPairFromSecretPhrase("foo")
	sm := util.NewSignedMessage(kp, m)

	fakeRequest := &Request {
		Message: sm,
		Response: nil,
	}

	s.ServeInBackground()
	s.requests <- fakeRequest
	// This hits the right code path but it feels like we ought to have a
	// better assertion here
	go s.Stop()
}

func sendString(address *Address, s string) error {
	c := NewClient(address)
	c.connect()
	c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	fmt.Fprintf(c.conn, s)
	_, err := util.ReadSignedMessage(c.conn)
	return err
}

func TestServerOkayWithMalformedMessage(t *testing.T) {
	_, configs := NewUnitTestNetwork()
	s := NewServer(configs[1])
	s.ServeInBackground()

	if sendString(s.LocalhostAddress(), "Garbage!\n\n\n") != io.EOF {
		t.Errorf("Didn't get disconnected after a total garbage message")
	}

	if sendString(s.LocalhostAddress(), "a:b:c:d\n") != io.EOF {
		t.Errorf("Didn't get disconnected after a semi-garbage message")
	}

	goodMessage := "{ \"T\": \"N\", \"M\": { \"I\": 1 } }"
	kp := util.NewKeyPair()
	line := fmt.Sprintf("e:%s:%s:%s\n",
		kp.PublicKey(), "notRealSignature", goodMessage)

	if sendString(s.LocalhostAddress(), line) != io.EOF {
		t.Errorf("Didn't get disconnected after a bad-signature message")
	}

	line = fmt.Sprintf("e:%s:%s:%s\n",
		kp.PublicKey(), kp.Sign(goodMessage), goodMessage)

	if sendString(s.LocalhostAddress(), line) != nil {
		t.Errorf("The server should still process a good message")
	}

	go s.Stop()
}
