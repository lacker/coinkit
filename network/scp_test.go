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

	a := amy.OutgoingMessage()
	bob.Handle("amy", a)
	cal.Handle("amy", a)
	dan.Handle("amy", a)

	// TODO: test something about the state here
}
