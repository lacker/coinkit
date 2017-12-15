package network

import (
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
	s := NewStateBuilder("foo", []string{"foo"}, 1)
	if !MeetsQuorum(s.nState, []string{"foo"}) {
		t.Fatal("foo should meet the quorum")
	}
	if MeetsQuorum(s.nState, []string{"bar"}) {
		t.Fatal("bar should not meet the quorum")
	}
}

func TestConsensus(t *testing.T) {
	members := []string{"amy", "bob", "cal", "dan"}
	amy := NewStateBuilder("amy", members, 3)
	bob := NewStateBuilder("bob", members, 3)
	cal := NewStateBuilder("cal", members, 3)
	dan := NewStateBuilder("dan", members, 3)

	// Let everyone receive an initial nomination from Amy
	a := amy.OutgoingMessage()
	bob.Handle("amy", a)
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
	// but still no candidates. This works even without dan, who has nothing accepted.
	b := bob.OutgoingMessage()
	amy.Handle("bob", b)
	cal.Handle("bob", b)
	c := cal.OutgoingMessage()
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
