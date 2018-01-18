package consensus

import (
	"log"
	"runtime/debug"
	"sort"
	"strings"
)

// The block chain is a sequence of slots, each with an int id.
// The consensus protocol works by figuring out, one by one, what SlotValue
// should go into the next open slot.
// In general, the way that two different nodes should come to an agreement
// on a slot value is by "combining" some proposed slot values.
// The meaning of "combine" is not specified in the protocol directly.
//
// For now each block just has a list of comments, and we combine by
// merging those lists.
// This isn't supposed to be useful, it's just for testing.
type SlotValue struct {
	Comments []string
}

func MakeSlotValue(comment string) SlotValue {
	return SlotValue{Comments: []string{comment}}
}

func Equal(a SlotValue, b SlotValue) bool {
	return strings.Join(a.Comments, ",") == strings.Join(b.Comments, ",")
}

// Combine is specific to what the slot values are
func Combine(a SlotValue, b SlotValue) SlotValue {
	cmap := make(map[string]bool)
	for _, c := range a.Comments {
		cmap[c] = true
	}
	for _, c := range b.Comments {
		cmap[c] = true
	}
	answer := []string{}
	for c, _ := range cmap {
		answer = append(answer, c)
	}
	sort.Strings(answer)
	return SlotValue{Comments: answer}
}

// CombineSlice just runs Combine on each value in a slice
func CombineSlice(list []SlotValue) SlotValue {
	if len(list) == 0 {
		panic("CombineSlice should not be called on empty slices")
	}
	if len(list) == 1 {
		return list[0]
	}
	mid := len(list) / 2
	return Combine(CombineSlice(list[:mid]), CombineSlice(list[mid:]))
}

func HasSlotValue(list []SlotValue, v SlotValue) bool {
	k := strings.Join(v.Comments, ",")
	for _, s := range list {
		if strings.Join(s.Comments, ",") == k {
			return true
		}
	}
	return false
}

func (v SlotValue) Hash() string {
	return HashString(strings.Join(v.Comments, ","))
}

func AssertNoDupes(list []SlotValue) {
	m := make(map[string]bool)
	for _, v := range list {
		s := strings.Join(v.Comments, ",")
		if m[s] {
			debug.PrintStack()
			log.Fatalf("dupe in %+v", list)
		}
		m[s] = true
	}
}
