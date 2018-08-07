package network

import (
	"fmt"
	"testing"
	"time"

	"github.com/lacker/coinkit/data"
	"github.com/lacker/coinkit/util"
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

type Fatalfer interface {
	Fatalf(format string, args ...interface{})
}

func makeServers(t Fatalfer) []*Server {
	config, kps := NewUnitTestNetwork()
	answer := []*Server{}
	for i, kp := range kps {
		db := data.NewTestDatabase(i)
		server := NewServer(kp, config, db)
		if server == nil {
			t.Fatalf("failed to construct server")
		}

		// In theory rebroadcasts should not be necessary unless we have node failures
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
	servers := makeServers(t)
	stopServers(servers)
	moreServers := makeServers(t)
	stopServers(moreServers)
}

// sendMoney waits until the operation clears
// it fatals if from doesn't have the money
func sendMoney(conn Connection, from *util.KeyPair, to *util.KeyPair, amount uint64) {
	account := GetAccount(conn, from.PublicKey().String())
	if account == nil || account.Balance < amount {
		util.Logger.Fatalf("%s did not have enough money", from.PublicKey().String())
	}
	seq := account.Sequence + 1
	operation := &data.SendOperation{
		Signer:   from.PublicKey().String(),
		Sequence: account.Sequence + 1,
		To:       to.PublicKey().String(),
		Amount:   amount,
		Fee:      0,
	}
	sop := data.NewSignedOperation(operation, from)
	om := data.NewOperationMessage(sop)
	sm := util.NewSignedMessage(om, from)
	conn.Send(sm)
	WaitToClear(conn, from.PublicKey().String(), seq)
}

// TODO: make sendMoney use sendOperation
func sendOperation(conn Connection, from *util.KeyPair, op *data.SignedOperation) {
	om := data.NewOperationMessage(op)
	sm := util.NewSignedMessage(om, from)
	conn.Send(sm)
	WaitToClear(conn, from.PublicKey().String(), op.GetSequence())
}

func TestSendMoney(t *testing.T) {
	servers := makeServers(t)
	start := time.Now()
	mint := util.NewKeyPairFromSecretPhrase("mint")
	bob := util.NewKeyPairFromSecretPhrase("bob")
	conn := NewRedialConnection(servers[0].LocalhostAddress(), nil)
	sendMoney(conn, mint, bob, 100)
	elapsed := time.Now().Sub(start).Seconds()
	if elapsed > 10.0 {
		t.Fatalf("sending money is too slow: %.2f seconds", elapsed)
	}
	stopServers(servers)
}

func makeConns(servers []*Server, n int) []Connection {
	conns := []Connection{}
	for {
		for _, server := range servers {
			conns = append(conns, NewRedialConnection(server.LocalhostAddress(), nil))
			if len(conns) == n {
				return conns
			}
		}
	}
}

// sendMoneyRepeatedly sends one unit of money repeat times and closes the done
// channel when it is done.
func sendMoneyRepeatedly(
	conn Connection, from *util.KeyPair, to *util.KeyPair, repeat int, done chan bool) {
	for i := 0; i < repeat; i++ {
		sendMoney(conn, from, to, 1)
	}
	close(done)
}

func benchmarkSendMoney(numConns int, b *testing.B) {
	servers := makeServers(b)
	conns := makeConns(servers, numConns)

	// Setup
	kps := []*util.KeyPair{}
	chans := []chan bool{}
	mint := util.NewKeyPairFromSecretPhrase("mint")
	for i := 0; i < numConns; i++ {
		kps = append(kps, util.NewKeyPairFromSecretPhrase(fmt.Sprintf("kp%d", i)))
		chans = append(chans, make(chan bool))
		for _, server := range servers {
			server.setBalance(kps[i].PublicKey().String(), uint64(b.N))
		}
	}
	b.ResetTimer()

	// Kickoff
	for i, conn := range conns {
		go sendMoneyRepeatedly(conn, kps[i], mint, b.N, chans[i])
	}

	// Wait for the finish
	for _, ch := range chans {
		<-ch
	}
	util.Logger.Printf("work is finished")
	for _, server := range servers {
		server.Stats()
	}

	// Clean up
	for _, conn := range conns {
		conn.Close()
	}
	stopServers(servers)
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
	config, kps := NewUnitTestNetwork()
	s := NewServer(kps[0], config, nil)

	m := &FakeMessage{Number: 4}
	kp := util.NewKeyPairFromSecretPhrase("foo")
	sm := util.NewSignedMessage(m, kp)

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

// TODO: reenable this test when the general db problems clear up
/*
func TestDataOperations(t *testing.T) {
	servers := makeServers(t)
	start := time.Now()
	mint := util.NewKeyPairFromSecretPhrase("mint")
	conn := NewRedialConnection(servers[0].LocalhostAddress(), nil)

	// Create a document
	op := data.MakeTestCreateOperation(1)
	sendOperation(conn, mint, op)

	// Query for the document.
	docs := FindDocuments(conn, map[string]interface{}{"foo": 1})
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc but got %d", len(docs))
	}
	ob := docs[0].Data
	// TODO: update this test once we don't dupe id into the regular data
	if ob.NumKeys() != 2 {
		t.Fatalf("expected two keys but got: %s", ob)
	}
	foo, ok := ob.GetInt("foo")
	if !ok {
		t.Fatalf("doc does not contain foo")
	}
	if foo != 1 {
		t.Fatalf("expected foo = 1 but got %d", foo)
	}

	elapsed := time.Now().Sub(start).Seconds()
	if elapsed > 10.0 {
		t.Fatalf("data operations are too slow: %.2f seconds", elapsed)
	}
	stopServers(servers)
}
*/
