package data

import (
	"testing"
)

func TestSendOperationProcessing(t *testing.T) {
	c := NewCache()
	payBob := &SendOperation{
		Sequence: 1,
		Amount:   100,
		Fee:      3,
		Signer:   "alice",
		To:       "bob",
	}
	if c.Validate(payBob) {
		t.Fatalf("alice should not be able to pay bob with no account")
	}
	c.SetBalance("alice", 50)
	if c.Validate(payBob) {
		t.Fatalf("alice should not be able to pay bob with only 50 money")
	}
	c.SetBalance("alice", 200)
	if !c.Validate(payBob) {
		t.Fatalf("alice should be able to pay bob with 200 money")
	}
	if !c.Process(payBob) {
		t.Fatalf("the payment should have worked")
	}
	if c.Validate(payBob) {
		t.Fatalf("validation should reject replay attacks")
	}
}

func TestReadThrough(t *testing.T) {
	db := NewTestDatabase(0)
	c1 := NewDatabaseCache(db, 1)
	a1 := c1.GetAccount("bob")
	if a1 != nil {
		t.Fatalf("expected nil account, got %+v", a1)
	}
	c2 := NewDatabaseCache(db, 1)
	a2 := &Account{
		Owner:    "bob",
		Sequence: 7,
		Balance:  100,
	}
	db.UpsertAccount(a2)
	db.Commit()
	a3 := c1.GetAccount("bob")
	if a3 != nil {
		t.Fatalf("expected c1 to not do read-through when cache is warm")
	}
	a4 := c2.GetAccount("bob")
	if a4 == nil || a4.Balance != 100 {
		t.Fatalf("bad a4: %+v", a4)
	}

	if c2.GetAccount("nonexistent") != nil {
		t.Fatalf("nonexistent existed")
	}
	prereads := db.reads
	if c2.GetAccount("nonexistent") != nil {
		t.Fatalf("nonexistent existed")
	}
	if prereads != db.reads {
		t.Fatalf("double nil read should not require a db hit")
	}
}

func TestWriteThrough(t *testing.T) {
	db := NewTestDatabase(0)
	c1 := NewDatabaseCache(db, 1)
	a1 := &Account{
		Owner:    "bob",
		Sequence: 8,
		Balance:  200,
	}
	c1.UpsertAccount(a1)
	db.Commit()
	c2 := NewDatabaseCache(db, 1)
	a2 := c2.GetAccount("bob")
	if a2 == nil || a2.Balance != 200 {
		t.Fatalf("writethrough fail: %+v", a2)
	}
}
