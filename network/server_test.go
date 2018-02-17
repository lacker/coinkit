package network

import (
	"fmt"
	"io"
	"log"
	"testing"
	"time"
	"net"

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

		// A high number essentially disables the rebroadcasts for these tests.
		// In theory they should not be necessary unless we have node failures
		// or lossy communication channels.
		server.RebroadcastInterval = 4 * time.Second

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
	account := client.GetAccount(from.PublicKey().String())
	if account == nil || account.Balance < amount {
		log.Fatalf("%s did not have enough money", from.PublicKey().String())
	}
	seq := account.Sequence + 1
	transaction := &currency.Transaction{
		From:     from.PublicKey().String(),
		Sequence: account.Sequence + 1,
		To:       to.PublicKey().String(),
		Amount:   amount,
		Fee:      0,
	}
	st := transaction.SignWith(from)
	tm := currency.NewTransactionMessage(st)
	sm := util.NewSignedMessage(from, tm)
	client.SendMessage(sm)
	client.WaitToClear(from.PublicKey().String(), seq)
}

func TestSendMoney(t *testing.T) {
	servers := makeServers()
	start := time.Now()
	mint := util.NewKeyPairFromSecretPhrase("mint")
	bob := util.NewKeyPairFromSecretPhrase("bob")
	client := NewClient(servers[0].LocalhostAddress())
	sendMoney(client, mint, bob, 100)
	log.Printf("transaction cleared")
	elapsed := time.Now().Sub(start).Seconds()
	if elapsed > 3.0 {
		t.Fatalf("sending money is too slow: %.2f seconds", elapsed)
	}
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
			server.SetBalance(kps[i].PublicKey().String(), uint64(b.N))
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
	for _, server := range servers {
		server.Stats()
	}
}

func BenchmarkSendMoney1(b *testing.B) {
	benchmarkSendMoney(1, b)
}

func BenchmarkSendMoney10(b *testing.B) {
	benchmarkSendMoney(10, b)
}

func BenchmarkSendMoney30(b *testing.B) {
	benchmarkSendMoney(30, b)
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

func checkForDeadSocket(c net.Conn) error {
	c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, err := util.ReadSignedMessage(c)

	// If our read timed out, let's try again until we get a definitive error.
	if err, ok := err.(net.Error); ok && err.Timeout() {
		 return checkForDeadSocket(c)
	}

	return err
}

func sendString(address *Address, s string) error {
	c := NewClient(address)
	c.connect()
	fmt.Fprintf(c.conn, s)
	return checkForDeadSocket(c.conn)
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
		kp.PublicKey().String(), "notRealSignature", goodMessage)

	if sendString(s.LocalhostAddress(), line) != io.EOF {
		t.Errorf("Didn't get disconnected after a bad-signature message")
	}

	line = fmt.Sprintf("e:%s:%s:%s\n",
		kp.PublicKey().String(), kp.Sign(goodMessage), goodMessage)

	if sendString(s.LocalhostAddress(), line) != nil {
		t.Errorf("The server should still process a good message")
	}

	go s.Stop()
}
