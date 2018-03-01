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
	err := db.SaveBlock(block)
	if err != nil {
		t.Fatal(err)
	}
	b2 := db.GetBlock(3)
	if b2.C != block.C {
		t.Fatal("block changed: %+v -> %+v", block, b2)
	}
}

func TestFetchNonexistentBlock(t *testing.T) {
	db := NewTestDatabase()
	b := db.GetBlock(4)
	if b != nil {
		t.Fatal("block should be nonexistent")
	}
}

func TestCantSaveTwice(t *testing.T) {
	db := NewTestDatabase()
	block := &Block{
		Slot:  4,
		Value: "hi",
		C:     1,
		H:     2,
	}
	err := db.SaveBlock(block)
	if err != nil {
		t.Fatal(err)
	}
	err = db.SaveBlock(block)
	if err == nil {
		t.Fatal("a block should not save twice")
	}
}

func TestLastBlock(t *testing.T) {
	db := NewTestDatabase()
	dropAll(db)
	db = NewTestDatabase()
	b := db.LastBlock()
	if b != nil {
		t.Fatal("expected last block nil but got %+v", b)
	}
	b = &Block{
		Slot:  5,
		Value: "five",
	}
	err := db.SaveBlock(b)
	if err != nil {
		t.Fatal(err)
	}
	b2 := db.LastBlock()
	if b2.Value != b.Value {
		t.Fatal("b2: %+v", b2)
	}
}

func dropAll(db *Database) {
	db.postgres.MustExec("DROP TABLE IF EXISTS blocks")
}

// Clean up both before and after running tests
func TestMain(m *testing.M) {
	db := NewTestDatabase()
	dropAll(db)
	answer := m.Run()
	dropAll(db)
	os.Exit(answer)
}

// TODO: test that:
// lastblock works
