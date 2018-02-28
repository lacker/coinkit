package data

import (
	"os"
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

func TestFetchNonexistentBlock(t *testing.T) {
	db := NewTestDatabase()
	b := db.GetBlock(100)
	if b != nil {
		t.Fatal("block should be nonexistent")
	}
}

func TestMain(m *testing.M) {
	answer := m.Run()
	db := NewTestDatabase()
	db.postgres.MustExec("DROP TABLE IF EXISTS blocks")
	os.Exit(answer)
}

// TODO: test that:
// block slots are enforced to be unique
// lastblock works
