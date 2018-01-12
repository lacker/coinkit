package network

import (
	"log"
	"math/rand"
	"testing"
)

// Simulate the sending of messages from source to target
func chainSend(source *Chain, target *Chain) {
	if source == target {
		return
	}
	messages := source.OutgoingMessages()
	for _, message := range messages {
		m := EncodeThenDecode(message)
		response := target.Handle(source.publicKey, m)
		if response != nil {
			r := EncodeThenDecode(response)
			x := source.Handle(target.publicKey, r)
			if x != nil {
				log.Fatal("infinite response loop")
			}
		}
	}
}

// Makes a cluster of chains that requires a consensus of more than two thirds.
func chainCluster(size int) []*Chain {
	qs, names := MakeTestQuorumSlice(size)
	chains := []*Chain{}
	for _, name := range names {
		chains = append(chains, NewEmptyChain(name, qs))
	}
	return chains
}

// checkProgress checks that blocks match up to and including limit.
// it errors if there is any disagreement in externalized values.
func checkProgress(chains []*Chain, limit int, t *testing.T) {
	first := chains[0]
	for i := 1; i < len(chains); i++ {
		chain := chains[i]
		for j := 1; j <= limit; j++ {
			// Check that this chain agrees with the first one for slot j
			blockValue := chain.history[j].external.X
			firstValue := first.history[j].external.X
			if !Equal(blockValue, firstValue) {
				log.Printf("%s externalized %+v for slot %d",
					first.publicKey, firstValue, j)
				log.Printf("%s externalized %+v for slot %d",
					chain.publicKey, blockValue, j)
				t.Fatal("this cannot be")
			}
		}
	}
}

// progress returns the number of blocks that all of these chains have externalized
func progress(chains []*Chain) int {
	minSlot := chains[0].current.slot
	for i := 1; i < len(chains); i++ {
		if chains[i].current.slot < minSlot {
			minSlot = chains[i].current.slot
		}
	}
	return minSlot - 1
}

func chainFuzzTest(chains []*Chain, seed int64, t *testing.T) {
	limit := 10
	rand.Seed(seed)
	log.Printf("fuzz testing chains with seed %d", seed)
	for i := 0; i < 10000; i++ {
		j := rand.Intn(len(chains))
		k := rand.Intn(len(chains))
		chainSend(chains[j], chains[k])
		if progress(chains) >= limit {
			break
		}
		if i%1000 == 0 {
			log.Printf("done round: %d ************************************", i)
		}
	}

	if progress(chains) < limit {
		LogChains(chains)
		t.Fatalf("with seed %d, we only externalized %d blocks",
			seed, progress(chains))
	}

	checkProgress(chains, 10, t)
}

func TestChainFullCluster(t *testing.T) {
	var i int64
	for i = 0; i < 10; i++ {
		c := chainCluster(4)
		chainFuzzTest(c, i, t)
	}
}
