package network

import (
	"log"
	"testing"
	"time"

	"coinkit/currency"
	"coinkit/util"
)

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

func TestNewServerCreatesSufficientPeers(t *testing.T) {
	c0 := NewLocalConfig(0)
	s0 := NewServer(c0)

	if (len(s0.peers) != NumPeers - 1) {
		t.Errorf("Didn't create the right number of peers %f %f", len(s0.peers), NumPeers - 1);
	}
}

func TestNewServerFailsIfPortTaken(t *testing.T) {
	s0 := NewServer(NewLocalConfig(0))
	s1 := NewServer(NewLocalConfig(0))

	go s0.ServeForever()
	err := s1.ServeForever()
	if (err == nil) {
		t.Errorf("Didn't error out when port is already in use")
	}
}

func TestServerOkayWithFakeWellFormattedMessage(t *testing.T) {
	s0 := NewServer(NewLocalConfig(0))

	m := &FakeMessage{Number: 4}
	kp := util.NewKeyPairFromSecretPhrase("foo")
	sm := util.NewSignedMessage(kp, m)

	fakeRequest := &Request {
		Message: sm,
		Response: nil,
	}

	go s0.ServeForever()
	s0.requests <- fakeRequest
	// This hits the right code path but it feels like we ought to have a real assertion here
}

func ResetConnectionAndSendString(c *Client, s string) {
	c.connected = false
	c.connect()
	c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	fmt.Fprintf(c.conn, s)
}

func TestServerOkayWithMalformedMessage(t *testing.T) {
	s := NewServer(NewLocalConfig(0))
	s.ServeForever()

	c := NewClient(s.port)

	ResetConnectionAndSendString(c, "Hello, I am sending you garbage.\n\n\n")
	_, err := util.ReadSignedMessage(c.conn)
	if (err != io.EOF) {
		t.Errorf("Didn't get disconnected after a malformed message")
	}

	ResetConnectionAndSendString(c, "a:b:c:d\n")
	_, err2 := util.ReadSignedMessage(c.conn)
	if (err2 != io.EOF) {
		t.Errorf("Didn't get disconnected after a malformed message")
	}

	goodMessage := "{ \"T\": \"N\", \"M\": { \"I\": 1 } }"
	kp := util.NewKeyPair()
	ResetConnectionAndSendString(c,
		fmt.Sprintf("e:%s:%s:%s\n", kp.PublicKey(), "notRealSignature", goodMessage))

	_, err3 := util.ReadSignedMessage(c.conn)
	if (err3 != io.EOF) {
		t.Errorf("Didn't get disconnected after a malformed message")
	}

	// Let's test that the server is still in good shape and can process a good message
	ResetConnectionAndSendString(c, fmt.Sprintf("e:%s:%s:%s\n", kp.PublicKey(), kp.Sign(goodMessage), goodMessage))
	_, err4 := util.ReadSignedMessage(c.conn)

	if (err4 != nil) {
		t.Errorf("Couldn't get a response after the good message")
	}
}
