package currency

import (
	"testing"
)

func TestTransactionProcessing(t *testing.T) {
	m := NewAccountMap()
	payBob := &Transaction{
		Sequence: 1,
		Amount: 100,
		Fee: 3,
		From: "alice",
		To: "bob",
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
