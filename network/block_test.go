package network

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestSolipsistQuorum(t *testing.T) {
	s := NewBlock("foo", MakeQuorumSlice([]string{"foo"}, 1), 1)
	if !MeetsQuorum(s.nState, []string{"foo"}) {
		t.Fatal("foo should meet the quorum")
	}
	if MeetsQuorum(s.nState, []string{"bar"}) {
		t.Fatal("bar should not meet the quorum")
	}
}

func TestNominationMessage(t *testing.T) {
	v := MakeSlotValue("hello")
	D := QuorumSlice{
		Members:   []string{"foo", "bar", "baz", "qux"},
		Threshold: 3,
	}
	m := &NominationMessage{
		I:   1,
		Nom: []SlotValue{v},
		Acc: []SlotValue{v},
		D:   D,
	}
	s := EncodeMessage(m)
	_, err := DecodeMessage(s)
	if err != nil {
		t.Fatal("could not decode message: %v", err)
	}
}

func TestConsensus(t *testing.T) {
	members := []string{"amy", "bob", "cal", "dan"}
	qs := MakeQuorumSlice(members, 3)
	amy := NewBlock("amy", qs, 1)
	bob := NewBlock("bob", qs, 1)
	cal := NewBlock("cal", qs, 1)
	dan := NewBlock("dan", qs, 1)

	// Let everyone receive an initial nomination from Amy
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

// Sum of received values for all the blocks
func rsum(blocks []*Block) int {
	answer := 0
	for _, block := range blocks {
		answer += block.nState.received
		answer += block.bState.received
	}
	return answer
}

// Simulate the pending messages being sent from source to target
func send(source *Block, target *Block) {
	if source == target {
		return
	}
	messages := source.OutgoingMessages()
	for _, message := range messages {
		encoded := EncodeMessage(message)
		m, err := DecodeMessage(encoded)
		if err != nil {
			log.Fatal("decoding error:", err)
		}
		target.Handle(source.publicKey, m)
	}
}

// Have the blocks send messages back and forth until they are making no more
// progress
func converge(blocks []*Block) {
	i := 0
	for {
		i++
		log.Printf("Pass %d", i)
		initial := rsum(blocks)
		for _, source := range blocks {
			for _, target := range blocks {
				send(source, target)
			}
		}
		if rsum(blocks) == initial {
			break
		}
	}
}

// Makes a cluster that requires a consensus of more than two thirds.
func cluster(size int) []*Block {
	threshold := 2*size/3 + 1
	names := []string{}
	for i := 0; i < size; i++ {
		names = append(names, fmt.Sprintf("node%d", i))
	}
	blocks := []*Block{}
	qs := MakeQuorumSlice(names, threshold)
	for _, name := range names {
		block := NewBlock(name, qs, 1)
		blocks = append(blocks, block)
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

func fuzzTest(blocks []*Block, seed int64, t *testing.T) {
	rand.Seed(seed)
	log.Printf("fuzz testing with seed %d", seed)
	for i := 0; i < 10000; i++ {
		j := rand.Intn(len(blocks))
		k := rand.Intn(len(blocks))
		send(blocks[j], blocks[k])
		if allDone(blocks) {
			break
		}
		if i%1000 == 0 {
			log.Printf("done round: %d", i)
		}
	}

	if !nominationConverged(blocks) {
		for i := 0; i < len(blocks); i++ {
			log.Printf("--------------------------------------------------------------------------")
			if blocks[i].nState != nil {
				blocks[i].nState.Show()
			}
		}

		log.Printf("**************************************************************************")

		t.Fatalf("fuzz testing with seed %d, nomination did not converge", seed)
	}

	if !allDone(blocks) {
		for i := 0; i < len(blocks); i++ {
			log.Printf("--------------------------------------------------------------------------")
			if blocks[i].bState != nil {
				blocks[i].bState.Show()
			}
		}

		log.Printf("**************************************************************************")
		t.Fatalf("fuzz testing with seed %d, ballots did not converge", seed)
	}
}

func TestBasicConvergence(t *testing.T) {
	c := cluster(4)
	converge(c)
	assertDone(c, t)
}

// Worked up to 100k
func TestFullCluster(t *testing.T) {
	var i int64
	for i = 0; i < 100; i++ {
		c := cluster(4)
		fuzzTest(c, i, t)
	}
}

// Worked up to 100k
func TestOneNodeKnockedOut(t *testing.T) {
	var i int64
	for i = 0; i < 100; i++ {
		c := cluster(4)
		knockout := c[0:3]
		fuzzTest(knockout, i, t)
	}
}
