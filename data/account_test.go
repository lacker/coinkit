package data

import (
	"bytes"
	"testing"
)

func TestSendOperationProcessing(t *testing.T) {
	m := NewAccountMap()
	payBob := &SendOperation{
		Sequence: 1,
		Amount:   100,
		Fee:      3,
		Signer:   "alice",
		To:       "bob",
	}
	if m.Validate(payBob) {
		t.Fatalf("alice should not be able to pay bob with no account")
	}
	m.SetBalance("alice", 50)
	if m.Validate(payBob) {
		t.Fatalf("alice should not be able to pay bob with only 50 money")
	}
	m.SetBalance("alice", 200)
	if !m.Validate(payBob) {
		t.Fatalf("alice should be able to pay bob with 200 money")
	}
	if !m.Process(payBob) {
		t.Fatalf("the payment should have worked")
	}
	if m.Validate(payBob) {
		t.Fatalf("validation should reject replay attacks")
	}
}

func TestAccountHashing(t *testing.T) {
	a1 := &Account{Sequence: 1, Balance: 2}
	a2 := &Account{Sequence: 1, Balance: 20}
	if bytes.Equal(a1.Bytes(), a2.Bytes()) {
		t.Fatal("bytes should not be two-to-one")
	}
}
