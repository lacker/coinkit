package network

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	"coinkit/consensus"
	"coinkit/currency"
	"coinkit/util"
)

func sendNodeToNodeMessages(source *Node, target *Node, t *testing.T) {
	messages := source.OutgoingMessages()
	for _, message := range messages {
		m := util.EncodeThenDecode(message)
		response, ok := target.Handle(source.publicKey.String(), m)
		if ok {
			x, ok := source.Handle(target.publicKey.String(), response)
			if ok {
				log.Printf("initial message: %+v", message)
				log.Printf("response message: %+v", response)
				log.Printf("re-response message: %+v", x)
				t.Fatal("infinite response loop")
			}
		}
	}
}

func maxAccountBalance(nodes []*Node) uint64 {
	answer := uint64(0)
	for _, node := range nodes {
		b := node.queue.MaxBalance()
		if b > answer {
			answer = b
		}
	}
	return answer
}

func TestNodeCatchup(t *testing.T) {
	kp := util.NewKeyPairFromSecretPhrase("client")
	kp2 := util.NewKeyPairFromSecretPhrase("bob")
	qs, names := consensus.MakeTestQuorumSlice(4)
	nodes := []*Node{}
	for _, name := range names {
		node := NewNode(name, qs, nil)
		node.queue.SetBalance(kp.PublicKey().String(), 100)
		nodes = append(nodes, node)
	}

	// Run a few rounds with the first three nodes
	for round := 1; round <= 3; round++ {
		tr := &currency.Transaction{
			From:     kp.PublicKey().String(),
			Sequence: uint32(round),
			To:       kp2.PublicKey().String(),
			Amount:   1,
			Fee:      0,
		}
		ts := []*currency.SignedTransaction{tr.SignWith(kp)}
		m := currency.NewTransactionMessage(ts...)
		nodes[0].Handle(kp.PublicKey().String(), m)
		for i := 0; i < 10; i++ {
			sendNodeToNodeMessages(nodes[0], nodes[1], t)
			sendNodeToNodeMessages(nodes[0], nodes[2], t)
			sendNodeToNodeMessages(nodes[1], nodes[2], t)
			sendNodeToNodeMessages(nodes[1], nodes[0], t)
			sendNodeToNodeMessages(nodes[2], nodes[0], t)
			sendNodeToNodeMessages(nodes[2], nodes[1], t)
		}
		for i := 0; i <= 2; i++ {
			if nodes[i].Slot() != round+1 {
				t.Fatalf("nodes[%d] did not finish round %d", i, round)
			}
		}
	}

	// The last node should be able to catch up
	for i := 0; i < 10; i++ {
		sendNodeToNodeMessages(nodes[0], nodes[3], t)
		sendNodeToNodeMessages(nodes[3], nodes[0], t)
		sendNodeToNodeMessages(nodes[1], nodes[3], t)
		sendNodeToNodeMessages(nodes[3], nodes[2], t)
		sendNodeToNodeMessages(nodes[2], nodes[3], t)
		sendNodeToNodeMessages(nodes[3], nodes[2], t)
	}
	if nodes[3].Slot() != 4 {
		t.Fatalf("catchup failed")
	}
}

func nodeFuzzTest(seed int64, t *testing.T) {
	initialMoney := uint64(4)

	numClients := 5
	clients := []*util.KeyPair{}
	for i := 0; i < numClients; i++ {
		kp := util.NewKeyPairFromSecretPhrase(fmt.Sprintf("client%d", i))
		clients = append(clients, kp)
	}

	clientMessages := []*currency.TransactionMessage{}
	for i, client := range clients {
		neighbor := clients[(i+1)%len(clients)]

		// Each client attempts to send 1 money to their neighbor
		// with a fee of 1, many times.
		// This should always end up with everyone having 1 money.
		// Proof is left as an exercise to the reader :D
		ts := []*currency.SignedTransaction{}
		for seq := uint32(1); seq < uint32(initialMoney); seq++ {
			t := &currency.Transaction{
				From:     client.PublicKey().String(),
				Sequence: seq,
				To:       neighbor.PublicKey().String(),
				Amount:   1,
				Fee:      1,
			}
			ts = append(ts, t.SignWith(client))
		}
		m := currency.NewTransactionMessage(ts...)
		clientMessages = append(clientMessages, m)
	}

	// 4 nodes running on 3-out-of-4
	qs, names := consensus.MakeTestQuorumSlice(4)
	nodes := []*Node{}
	for _, name := range names {
		node := NewNode(name, qs, nil)
		for _, client := range clients {
			node.queue.SetBalance(client.PublicKey().String(), initialMoney)
		}
		nodes = append(nodes, node)
	}

	rand.Seed(seed ^ 789789)
	log.Printf("fuzz testing nodes with seed %d", seed)
	for i := 0; i <= 10000; i++ {
		if rand.Intn(2) == 0 {
			// Pick a random pair of nodes to exchange messages
			source := nodes[rand.Intn(len(nodes))]
			target := nodes[rand.Intn(len(nodes))]
			sendNodeToNodeMessages(source, target, t)
		} else {
			// Send a client-to-node message
			j := rand.Intn(len(clientMessages))
			client := clients[j]
			m := clientMessages[j]
			node := nodes[rand.Intn(len(nodes))]
			node.Handle(client.PublicKey().String(), m)
		}

		// Check if we are done
		if maxAccountBalance(nodes) == 1 {
			break
		}
	}

	if maxAccountBalance(nodes) != 1 {
		for _, node := range nodes {
			node.Log()
		}
		t.Fatalf("failure to converge with seed %d", seed)
	}
}

// Works up to 1k
func TestNodeFullCluster(t *testing.T) {
	var i int64
	for i = 1; i <= util.GetTestLoopLength(2, 1000); i++ {
		nodeFuzzTest(i, t)
	}
}
