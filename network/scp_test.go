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
		t.Fatal("e is", e)
	}
}
