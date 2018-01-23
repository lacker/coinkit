package currency

import (
	"encoding/json"
	"log"
	
	"coinkit/util"
)

type Transaction struct {
	// Who is sending this money
	From string
	
	// The sequence number for this transaction
	Sequence uint32

	// Who is receiving this money
	To string
	
	// The amount of currency to transfer
	Amount uint64

	// How much the sender is willing to pay to get this transfer registered
	// This is on top of the amount
	Fee uint64
}

type SignedTransaction struct {
	*Transaction

	// The signature to prove that the sender has signed this
	// Nil if the transaction has not been signed
	Signature string	
}

// Signs the transaction with the provided keypair.
// The caller must check the keypair is the actual sender.
func (t *Transaction) SignWith(keyPair *util.KeyPair) *SignedTransaction {
	if keyPair.PublicKey() != t.From {
		panic("you can only sign your own transactions")
	}
	bytes, err := json.Marshal(t)
	if err != nil {
		panic("failed to sign transaction because json encoding failed")
	}
	return &SignedTransaction{
		Transaction: t,
		Signature: keyPair.Sign(string(bytes)),
	}
}

func (s *SignedTransaction) Verify() bool {
	if s.Transaction == nil {
		return false
	}
	bytes, err := json.Marshal(s.Transaction)
	if err != nil {
		return false
	}
	return util.Verify(s.Transaction.From, string(bytes), s.Signature)
}

// Sort message so they are higher fees first.
func (t *Transaction) OrderedBefore(other *Transaction) bool {
	if other == nil {
		log.Fatal("cannot compare nil transaction message")
	}
	if t.Fee > other.Fee {
		return true
	}
	if t.Fee < other.Fee {
		return false
	}

	// Ties are ok
	return false
}

func (s *SignedTransaction) OrderedBefore(other *SignedTransaction) bool {
	return s.Transaction.OrderedBefore(other.Transaction)
}
