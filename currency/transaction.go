package currency

import (
	"encoding/json"
	"fmt"
	"strings"
	
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

func (t *Transaction) String() string {
	return fmt.Sprintf("send %d from %s -> %s, seq %d fee %d",
		t.Amount, util.Shorten(t.From), util.Shorten(t.To), t.Sequence, t.Fee)
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

// HighestPriorityFirst is a comparator in the emirpasic/gods comparator style.
// Negative return indicates a < b
// Positive return indicates a > b
// Comparison indicates overall "priority" putting the highest priority first.
// This means that when a has a higher fee than b, a < b.
func HighestPriorityFirst(a, b interface{}) int {
	s1 := a.(*SignedTransaction)
	s2 := b.(*SignedTransaction)

	switch {
	case s1.Transaction.Fee > s2.Transaction.Fee:
		// s1 is higher priority. so a < b
		return -1
	case s1.Transaction.Fee < s2.Transaction.Fee:
		return 1
	case s1.Signature < s2.Signature:
		// s1 is higher priority
		return -1
	case s1.Signature > s2.Signature:
		return 1
	default:
		return 0
	}
}

func makeTestTransaction(n int) *SignedTransaction {
	kp := util.NewKeyPairFromSecretPhrase(fmt.Sprintf("blorp %d", n))
	t := &Transaction{
		From: kp.PublicKey(),
		Sequence: 1,
		To: "nobody",
		Amount: uint64(n),
		Fee: uint64(n),
	}
	return t.SignWith(kp)
}

func StringifyTransactions(transactions []*SignedTransaction) string {
	parts := []string{}
	limit := 2
	for i, t := range transactions {
		if i >= limit {
			parts = append(parts, fmt.Sprintf("and %d more",
				len(transactions) - limit))
			break
		}
		parts = append(parts, t.String())
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, "; "))
}

