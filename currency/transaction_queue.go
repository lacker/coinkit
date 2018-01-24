package currency

import (
	"github.com/emirpasic/gods/sets/treeset"
)

// QueueLimit defines how many items will be held in the queue at a time
const QueueLimit = 1000

// SharingSize tells how many queue items get shared by default
const SharingSize = 100

type TransactionQueue struct {
	set *treeset.Set
}

func NewTransactionQueue() *TransactionQueue {
	return &TransactionQueue{
		set: treeset.NewWith(HighestPriorityFirst),
	}
}

// Returns the top n items in the queue
// If the queue does not have enough, return as many as we can
func (q *TransactionQueue) Top(n int) []*SignedTransaction {
	answer := []*SignedTransaction{}
	for _, item := range q.set.Values() {
		answer = append(answer, item.(*SignedTransaction))
		if len(answer) == n {
			break
		}
	}
	return answer
}

// Remove removes a transaction from the queue
func (q *TransactionQueue) Remove(t *SignedTransaction) {
	if t == nil {
		return
	}
	q.set.Remove(t)
}

// Add adds a transaction to the queue
// If it can't be verified, we discard it.
func (q *TransactionQueue) Add(t *SignedTransaction) {
	if t == nil || !t.Verify() {
		return
	}
	q.set.Add(t)
	if q.set.Size() > QueueLimit {
		it := q.set.Iterator()
		if !it.Last() {
			panic("logical failure with treeset")
		}
		q.set.Remove(it.Value())
	}
}

// SharingMessage returns a message that other transaction queues can use
// to have the same transactions that we do
func (q *TransactionQueue) SharingMessage() *TransactionMessage {
	return &TransactionMessage{
		Transactions: q.Top(SharingSize),
	}
}

func (q *TransactionQueue) Size() int {
	return q.set.Size()
}
