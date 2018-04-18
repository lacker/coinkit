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
	// TODO: implement
}
