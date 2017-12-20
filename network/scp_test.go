package network

import (
	"fmt"
	"log"
	"strings"
	"testing"
)

func TestCombineSlotValues(t *testing.T) {
	a := MakeSlotValue("foo")
	b := MakeSlotValue("bar")
	c := MakeSlotValue("baz")
	d := Combine(a, b)
	e := Combine(d, c)
	if strings.Join(e.Comments, ",") != "bar,baz,foo" {
		t.Fatal("a is", a)
		t.Fatal("e is", e)
	}
}

func TestSolipsistQuorum(t *testing.T) {
	s := NewChainState("foo", []string{"foo"}, 1)
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
		Members: []string{"foo", "bar", "baz", "qux"},
		Threshold: 3,
	}
	m := &NominationMessage{
		I: 1,
		X: []SlotValue{v},
		Y: []SlotValue{v},
		D: D,
	}
	s := EncodeMessage(m)
	_, err := DecodeMessage(s)
	if err != nil {
		t.Fatal("could not decode message: %v", err)
	}
}

func TestConsensus(t *testing.T) {
	members := []string{"amy", "bob", "cal", "dan"}
	amy := NewChainState("amy", members, 3)
	bob := NewChainState("bob", members, 3)
	cal := NewChainState("cal", members, 3)
	dan := NewChainState("dan", members, 3)

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

// Sum of received values for all the chains
func rsum(chains []*ChainState) int {
	answer := 0
	for _, chain := range chains {
		answer += chain.nState.received
		answer += chain.bState.received
	}
	return answer
}

// Have the chains send messages back and forth until they are making no more
// progress
func converge(chains []*ChainState) {
	for {
		initial := rsum(chains)
		for _, chain := range chains {
			messages := chain.OutgoingMessages()
			for _, message := range messages {
				encoded := EncodeMessage(message)
				m, err := DecodeMessage(encoded)
				if err != nil {
					log.Fatal("decoding error:", err)
				}
				for _, target := range chains {
					if chain != target {
						target.Handle(chain.publicKey, m)
					}
				}
			}
		}
		if rsum(chains) == initial {
			break
		}
	}
}

// Makes a cluster that requires a consensus of more than two thirds.
func cluster(size int) []*ChainState {
	threshold := 2 * size / 3 + 1
	names := []string{}
	for i := 0; i < size; i++ {
		names = append(names, fmt.Sprintf("node%d", i))
	}
	chains := []*ChainState{}
	for _, name := range names {
		chain := NewChainState(name, names, threshold)
		chains = append(chains, chain)
	}
	return chains
}

func TestConvergence(t *testing.T) {
	c := cluster(4)
	converge(c)
}
