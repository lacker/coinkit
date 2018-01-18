package consensus

import (
	"log"
	"math/rand"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestSolipsistQuorum(t *testing.T) {
	s := NewBlock("foo", MakeQuorumSlice([]string{"foo"}, 1), 1, "")
	if !MeetsQuorum(s.nState, []string{"foo"}) {
		t.Fatal("foo should meet the quorum")
	}
	if MeetsQuorum(s.nState, []string{"bar"}) {
		t.Fatal("bar should not meet the quorum")
	}
}

func TestConsensus(t *testing.T) {
	members := []string{"amy", "bob", "cal", "dan"}
	qs := MakeQuorumSlice(members, 3)
	amy := NewBlock("amy", qs, 1, "")
	bob := NewBlock("bob", qs, 1, "")
	cal := NewBlock("cal", qs, 1, "")
	dan := NewBlock("dan", qs, 1, "")

	// Let everyone receive an initial nomination from Amy
	amy.nState.NominateNewValue(MakeSlotValue("hello its amy"))
	a := amy.OutgoingMessages()[0]
	
	bob.Handle("amy", a)
	if len(bob.nState.N) != 1 {
		t.Fatal("len(bob.nState.N) != 1")
	}
	cal.Handle("amy", a)
	dan.Handle("amy", a)

	// At this point everyone should have a nomination
	if !amy.nState.HasNomination() {
		t.Fatal("!amy.nState.HasNomination()")
	}
	if !bob.nState.HasNomination() {
		t.Fatal("!bob.nState.HasNomination()")
	}
	if !cal.nState.HasNomination() {
		t.Fatal("!cal.nState.HasNomination()")
	}
	if !dan.nState.HasNomination() {
		t.Fatal("!dan.nState.HasNomination()")
	}

	// Once bob and cal broadcast, everyone should have one accepted value,
	// but still no candidates. This works even without dan, who has nothing
	// accepted.
	b := bob.OutgoingMessages()[0]
	amy.Handle("bob", b)
	if len(amy.nState.N) != 1 {
		t.Fatalf("amy.nState.N = %#v", amy.nState.N)
	}
	cal.Handle("bob", b)
	c := cal.OutgoingMessages()[0]
	amy.Handle("cal", c)
	bob.Handle("cal", c)
	if len(amy.nState.Y) != 1 {
		t.Fatal("len(amy.nState.Y) != 1")
	}
	if len(bob.nState.Y) != 1 {
		t.Fatal("len(bob.nState.Y) != 1")
	}
	if len(cal.nState.Y) != 1 {
		t.Fatal("len(cal.nState.Y) != 1")
	}
	if len(dan.nState.Y) != 0 {
		t.Fatal("len(dan.nState.Y) != 0")
	}
}

// Simulate the pending messages being sent from source to target
func blockSend(source *Block, target *Block) {
	if source == target {
		return
	}
	messages := source.OutgoingMessages()
	for _, message := range messages {
		target.Handle(source.publicKey, message)
	}
}

// Makes a cluster that requires a consensus of more than two thirds.
func blockCluster(size int) []*Block {
	qs, names := MakeTestQuorumSlice(size)
	blocks := []*Block{}
	for _, name := range names {
		blocks = append(blocks, NewBlock(name, qs, 1, ""))
	}
	return blocks
}

func allDone(blocks []*Block) bool {
	for _, block := range blocks {
		if !block.Done() {
			return false
		}
	}
	return true
}

// assertDone verifies that every block has gotten to externalize
func assertDone(blocks []*Block, t *testing.T) {
	for _, block := range blocks {
		if !block.Done() {
			t.Fatalf("%s is not externalizing: %s",
				block.publicKey, spew.Sdump(block))
		}
	}
}

func nominationConverged(blocks []*Block) bool {
	var value SlotValue
	for i, block := range blocks {
		if !block.nState.HasNomination() {
			return false
		}
		if i == 0 {
			value = block.nState.PredictValue()
		} else {
			v := block.nState.PredictValue()
			if !Equal(value, v) {
				return false
			}
		}
	}
	return true
}

func blockFuzzTest(blocks []*Block, seed int64, t *testing.T) {
	rand.Seed(seed ^ 1234569)
	log.Printf("fuzz testing blocks with seed %d", seed)
	for i := 0; i < 10000; i++ {
		j := rand.Intn(len(blocks))
		k := rand.Intn(len(blocks))
		blockSend(blocks[j], blocks[k])
		if allDone(blocks) {
			return
		}
		if i%1000 == 0 {
			log.Printf("done round: %d", i)
		}
	}

	if !nominationConverged(blocks) {
		log.Printf("nomination did not converge")
		for i := 0; i < len(blocks); i++ {
			log.Printf("--------------------------------------------------------------------------")
			if blocks[i].nState != nil {
				blocks[i].nState.Show()
			}
		}

		log.Printf("**************************************************************************")

		t.Fatalf("fuzz testing with seed %d, nomination did not converge", seed)
	}

	log.Printf("balloting did not converge")		
	for i := 0; i < len(blocks); i++ {
		log.Printf("--------------------------------------------------------------------------")
		if blocks[i].bState != nil {
			blocks[i].bState.Show()
		}
	}

	log.Printf("**************************************************************************")
	t.Fatalf("fuzz testing with seed %d, ballots did not converge", seed)
}

// Should work to 100k
func TestBlockFullCluster(t *testing.T) {
	var i int64
	for i = 0; i < 100; i++ {
		c := blockCluster(4)
		blockFuzzTest(c, i, t)
	}
}

// Should work to 100k
func TestBlockOneNodeKnockedOut(t *testing.T) {
	var i int64
	for i = 0; i < 100; i++ {
		c := blockCluster(4)
		knockout := c[0:3]
		blockFuzzTest(knockout, i, t)
	}
}
