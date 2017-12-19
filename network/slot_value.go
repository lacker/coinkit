package network

import (
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
	joined := append(a.Comments, b.Comments...)
	sort.Strings(joined)
	answer := []string{}
	for _, item := range joined {
		if len(answer) == 0 || answer[len(answer)-1] != item {
			answer = append(answer, item)
		}
	}
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

