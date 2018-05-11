package data

import (
	"testing"

	"github.com/lacker/coinkit/util"
)

func TestFullQueue(t *testing.T) {
	kp := util.NewKeyPair()
	q := NewOperationQueue(kp.PublicKey(), nil, 1)
	for i := 1; i <= QueueLimit+10; i++ {
		op := makeTestSendOperation(i)
		send := op.Operation.(*SendOperation)
		q.cache.SetBalance(send.Signer, 10*send.Amount)
		if !op.Verify() {
			t.Fatalf("bad op: %+v", op)
		}

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
	q.Add(&SignedOperation{})
	if q.Size() != 0 {
		t.Fatalf("queue should be empty")
	}
}

func TestSendOperation(t *testing.T) {
	kp := util.NewKeyPair()
	q := NewOperationQueue(kp.PublicKey(), nil, 1)
	if q.OperationMessage() != nil {
		t.Fatal("there should be no operation message with an empty queue")
	}
	op := makeTestSendOperation(0)
	tr := op.Operation.(*SendOperation)
	q.cache.SetBalance(tr.Signer, 10*tr.Amount)
	if !op.Verify() {
		t.Fatal("bad op")
	}
	q.Add(op)
	if q.OperationMessage() == nil {
		t.Fatal("there should be an operation message after we add a send operation")
	}
}

func TestCreateOperation(t *testing.T) {
	kp := util.NewKeyPair()
	q := NewOperationQueue(kp.PublicKey(), nil, 1)
	op := makeTestCreateOperation(1)
	if !op.Verify() {
		t.Fatal("bad op")
	}
	q.Add(op)
	if q.OperationMessage() != nil {
		t.Fatal("the op should be invalid because of insufficient balance")
	}
	q.cache.SetBalance(op.GetSigner(), TotalMoney)
	q.Add(op)
	if q.OperationMessage() == nil {
		t.Fatal("there should be an op message with a create operation")
	}
}
