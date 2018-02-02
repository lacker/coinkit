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

func (m *FakeMessage) String() string {
	return "Fake"
}

func makeServers() []*Server {
	_, configs := NewUnitTestNetwork()
	answer := []*Server{}
	for _, config := range configs {
		server := NewServer(config)
		server.InitMint()

		// Essentially disable the rebroadcasts for these tests.
		// In theory they should not be necessary unless we have node failures
		// or lossy communication channels.
		server.RebroadcastInterval = 60 * time.Second

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

// sendMoney waits until the transaction clears
// it fatals if from doesn't have the money
func sendMoney(client *Client, from *util.KeyPair, to *util.KeyPair, amount uint64) {
	account := client.GetAccount(from.PublicKey())
	if account == nil || account.Balance < amount {
		log.Fatalf("%s did not have enough money", from.PublicKey())
	}
	seq := account.Sequence + 1
	transaction := &currency.Transaction{
		From:     from.PublicKey(),
		Sequence: account.Sequence + 1,
		To:       to.PublicKey(),
		Amount:   amount,
		Fee:      0,
	}
	st := transaction.SignWith(from)
	tm := currency.NewTransactionMessage(st)
	sm := util.NewSignedMessage(from, tm)
	client.SendMessage(sm)
	client.WaitToClear(from.PublicKey(), seq)
}

func TestSendMoney(t *testing.T) {
	servers := makeServers()
	mint := util.NewKeyPairFromSecretPhrase("mint")
	bob := util.NewKeyPairFromSecretPhrase("bob")
	client := NewClient(servers[0].LocalhostAddress())
	sendMoney(client, mint, bob, 100)
	log.Printf("transaction cleared")
	go stopServers(servers)
}

func makeClients(servers []*Server, n int) []*Client {
	clients := []*Client{}
	for {
		for _, server := range servers {
			clients = append(clients, NewClient(server.LocalhostAddress()))
			if len(clients) == n {
				return clients
			}
		}
	}
}

// sendMoneyRepeatedly sends one unit of money repeat times and closes the done
// channel when it is done.
func sendMoneyRepeatedly(
	client *Client, from *util.KeyPair, to *util.KeyPair, repeat int, done chan bool) {
	for i := 0; i < repeat; i++ {
		sendMoney(client, from, to, 1)
	}
	close(done)
}

func benchmarkSendMoney(numClients int, b *testing.B) {
	servers := makeServers()
	clients := makeClients(servers, numClients)

	// Setup
	kps := []*util.KeyPair{}
	chans := []chan bool{}
	mint := util.NewKeyPairFromSecretPhrase("mint")
	for i := 0; i < numClients; i++ {
		kps = append(kps, util.NewKeyPairFromSecretPhrase(fmt.Sprintf("kp%d", i)))
		chans = append(chans, make(chan bool))
		for _, server := range servers {
			server.SetBalance(kps[i].PublicKey(), uint64(b.N))
		}
	}
	b.ResetTimer()

	// Kickoff
	for i, client := range clients {
		go sendMoneyRepeatedly(client, kps[i], mint, b.N, chans[i])
	}

	// Wait for the finish
	for _, ch := range chans {
		<-ch
	}
}

func BenchmarkSendMoney10(b *testing.B) {
	benchmarkSendMoney(10, b)
}

func TestServerOkayWithFakeWellFormattedMessage(t *testing.T) {
	_, configs := NewUnitTestNetwork()
	s := NewServer(configs[0])

	m := &FakeMessage{Number: 4}
	kp := util.NewKeyPairFromSecretPhrase("foo")
	sm := util.NewSignedMessage(kp, m)

	fakeRequest := &Request{
		Message:  sm,
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
