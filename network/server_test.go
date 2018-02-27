package network

import (
	"fmt"
	"log"
	"testing"
	"time"

	"coinkit/currency"
	"coinkit/util"
)

func makeServers() []*Server {
	_, configs := NewUnitTestNetwork()
	answer := []*Server{}
	for _, config := range configs {
		server := NewServer(config)
		server.InitMint()

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
	servers := makeServers()
	stopServers(servers)
	moreServers := makeServers()
	stopServers(moreServers)
}

// sendMoney waits until the transaction clears
// it fatals if from doesn't have the money
func sendMoney(conn Connection, from *util.KeyPair, to *util.KeyPair, amount uint64) {
	account := GetAccount(conn, from.PublicKey().String())
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
	conn.Send(sm)
	WaitToClear(conn, from.PublicKey().String(), seq)
}

func TestSendMoney(t *testing.T) {
	servers := makeServers()
	start := time.Now()
	mint := util.NewKeyPairFromSecretPhrase("mint")
	bob := util.NewKeyPairFromSecretPhrase("bob")
	conn := NewRedialConnection(servers[0].LocalhostAddress(), nil)
	sendMoney(conn, mint, bob, 100)
	elapsed := time.Now().Sub(start).Seconds()
	if elapsed > 3.0 {
		t.Fatalf("sending money is too slow: %.2f seconds", elapsed)
	}
	go stopServers(servers)
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
	servers := makeServers()
	conns := makeConns(servers, numConns)

	// Setup
	kps := []*util.KeyPair{}
	chans := []chan bool{}
	mint := util.NewKeyPairFromSecretPhrase("mint")
	for i := 0; i < numConns; i++ {
		kps = append(kps, util.NewKeyPairFromSecretPhrase(fmt.Sprintf("kp%d", i)))
		chans = append(chans, make(chan bool))
		for _, server := range servers {
			server.SetBalance(kps[i].PublicKey().String(), uint64(b.N))
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
	log.Printf("work is finished")
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
