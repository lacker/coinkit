package currency

import (
	"fmt"
	"testing"

	"coinkit/util"
)

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

func TestFullQueue(t *testing.T) {
	q := NewTransactionQueue()
	for i := 1; i <= QueueLimit + 10; i++ {
		t := makeTestTransaction(i)
		q.accounts.SetBalance(t.Transaction.From, 10 * t.Transaction.Amount)
		q.Add(t)
	}
	if q.Size() != QueueLimit {
		t.Fatalf("q.Size() was %d", q.Size())
	}
	top := q.Top(11)
	if top[10].Transaction.Amount != QueueLimit {
		t.Fatalf("top is wrong")
	}
	for i := 1; i <= QueueLimit + 10; i++ {
		q.Remove(makeTestTransaction(i))
	}
	q.Add(nil)
	q.Add(&SignedTransaction{})
	if q.Size() != 0 {
		t.Fatalf("queue should be empty")
	}
}
