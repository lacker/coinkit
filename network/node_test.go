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
		response := target.Handle(source.publicKey, m)
		if response != nil {
			x := source.Handle(target.publicKey, response)
			if x != nil {
				log.Fatal("infinite response loop")
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
		neighbor := clients[(i+1) % len(clients)]

		// Each client attempts to send 1 money to their neighbor
		// with a fee of 1, many times.
		// This should always end up with everyone having 1 money.
		// Proof is left as an exercise to the reader :D		
		ts := []*currency.SignedTransaction{}
		for seq := uint32(1); seq < uint32(initialMoney); seq++ {
			t := &currency.Transaction{
				From: client.PublicKey(),
				Sequence: seq,
				To: neighbor.PublicKey(),
				Amount: 1,
				Fee: 1,
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
		node := NewNode(name, qs)
		for _, client := range clients {
			node.queue.SetBalance(client.PublicKey(), initialMoney)
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
			node.Handle(client.PublicKey(), m)
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
