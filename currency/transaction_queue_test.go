package currency

import (
	"testing"

	"github.com/lacker/coinkit/util"
)

func TestFullQueue(t *testing.T) {
	kp := util.NewKeyPair()
	q := NewTransactionQueue(kp.PublicKey())
	for i := 1; i <= QueueLimit+10; i++ {
		op := makeTestSendOperation(i)
		t := op.Operation.(*SendOperation)
		q.accounts.SetBalance(t.Signer, 10*t.Amount)
		q.Add(op)
	}
	if q.Size() != QueueLimit {
		t.Fatalf("q.Size() was %d", q.Size())
	}
	top := q.Top(11)
	if top[10].Operation.(*SendOperation).Amount != QueueLimit {
		t.Fatalf("top is wrong")
	}
	for i := 1; i <= QueueLimit+10; i++ {
		q.Remove(makeTestSendOperation(i))
	}
	q.Add(nil)
	q.Add(&util.SignedOperation{})
	if q.Size() != 0 {
		t.Fatalf("queue should be empty")
	}
}

func TestTransactionMessage(t *testing.T) {
	kp := util.NewKeyPair()
	q := NewTransactionQueue(kp.PublicKey())
	if q.TransactionMessage() != nil {
		t.Fatal("there should be no transaction message with an empty queue")
	}
	op := makeTestSendOperation(0)
	tr := op.Operation.(*SendOperation)
	q.accounts.SetBalance(tr.Signer, 10*tr.Amount)
	q.Add(op)
	if q.TransactionMessage() == nil {
		t.Fatal("there should be a transaction message after we add one operation")
	}
}
