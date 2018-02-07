package currency

import (
	"testing"

	"coinkit/util"
)

func TestFullQueue(t *testing.T) {
	kp := util.NewKeyPair()
	q := NewTransactionQueue(kp.PublicKey())
	for i := 1; i <= QueueLimit+10; i++ {
		t := makeTestTransaction(i)
		q.accounts.SetBalance(t.Transaction.From, 10*t.Transaction.Amount)
		q.Add(t)
	}
	if q.Size() != QueueLimit {
		t.Fatalf("q.Size() was %d", q.Size())
	}
	top := q.Top(11)
	if top[10].Transaction.Amount != QueueLimit {
		t.Fatalf("top is wrong")
	}
	for i := 1; i <= QueueLimit+10; i++ {
		q.Remove(makeTestTransaction(i))
	}
	q.Add(nil)
	q.Add(&SignedTransaction{})
	if q.Size() != 0 {
		t.Fatalf("queue should be empty")
	}
}

func TestSharingMessage(t *testing.T) {
	kp := util.NewKeyPair()
	q := NewTransactionQueue(kp.PublicKey())
	if q.SharingMessage() != nil {
		t.Fatal("there should be no sharing message with an empty queue")
	}
	tr := makeTestTransaction(0)
	q.accounts.SetBalance(tr.Transaction.From, 10*tr.Transaction.Amount)
	q.Add(tr)
	if q.SharingMessage() == nil {
		t.Fatal("there should be a sharing message after we add one transaction")
	}
}
