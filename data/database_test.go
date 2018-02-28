package data

import (
	"testing"
)

func TestSaveAndFetch(t *testing.T) {
	db := NewTestDatabase()
	block := &Block{
		Slot:  3,
		Value: "foo",
		C:     3,
		H:     4,
	}
	db.SaveBlock(block)
	b2 := db.GetBlock(3)
	if b2.C != block.C {
		t.Fatal("block changed: %+v -> %+v", block, b2)
	}
}

// TODO: test that:
// saving and fetching works ok
// fetching a nonexistent block does not die
// block slots are enforced to be unique
// lastblock works
