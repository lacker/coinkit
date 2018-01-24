package currency

import (
	"log"
	
	"github.com/emirpasic/gods/sets/treeset"
)

// MaxChunkSize defines how many items can be put in a chunk
const MaxChunkSize = 100

// QueueLimit defines how many items will be held in the queue at a time
const QueueLimit = 1000

// TransactionQueue keeps the transactions that are pending but have neither
// been rejected nor confirmed.
type TransactionQueue struct {
	set *treeset.Set

	// Transactions that we have not yet shared
	outbox []*SignedTransaction

	// accounts is used to validate transactions
	// For now this is the actual authentic store of account data
	// TODO: get this into a real database
	accounts *AccountMap
}

func NewTransactionQueue() *TransactionQueue {
	return &TransactionQueue{
		set: treeset.NewWith(HighestPriorityFirst),
		outbox: []*SignedTransaction{},
		accounts: NewAccountMap(),
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
// If it isn't valid, we just discard it.
// We don't constantly revalidate so it's possible we have invalid
// transactions in the queue.
func (q *TransactionQueue) Add(t *SignedTransaction) {
	if !q.Validate(t) {
		return
	}
	q.set.Add(t)
	if q.set.Size() > QueueLimit {
		it := q.set.Iterator()
		if !it.Last() {
			log.Fatal("logical failure with treeset")
		}
		worst := it.Value()
		q.set.Remove(worst)
	}
}

func (q *TransactionQueue) Handle(m *TransactionMessage) {
	if m == nil {
		return
	}
	for _, t := range m.Transactions {
		q.Add(t)
	}
}

func (q *TransactionQueue) Contains(t *SignedTransaction) bool {
	return q.set.Contains(t)
}

func (q *TransactionQueue) Values() []*SignedTransaction {
	answer := []*SignedTransaction{}
	for _, t := range q.set.Values() {
		answer = append(answer, t.(*SignedTransaction))
	}
	return answer
}

// SharingMessage returns the messages we want to share with other nodes.
// We only want to share once, so this does mutate the queue.
// Returns nil if we have nothing to share.
func (q *TransactionQueue) SharingMessage() *TransactionMessage {
	ts := []*SignedTransaction{}
	for _, t := range q.outbox {
		if q.Contains(t) {
			ts = append(ts, t)
		}
	}
	q.outbox = []*SignedTransaction{}
	if len(ts) == 0 {
		return nil
	}
	return &TransactionMessage{
		Transactions: ts,
	}
}

func (q *TransactionQueue) Size() int {
	return q.set.Size()
}

func (q *TransactionQueue) Validate(t *SignedTransaction) bool {
	return t != nil && t.Verify() && q.accounts.Validate(t.Transaction)
}

func (q *TransactionQueue) ValidateChunk(chunk *LedgerChunk) bool {
	if chunk == nil {
		return false
	}

	if len(chunk.Transactions) > MaxChunkSize {
		return false
	}
	
	validator := q.accounts.CowCopy()
	for _, t := range chunk.Transactions {
		if !q.Validate(t) || !validator.Process(t.Transaction) {
			return false
		}
	}

	for owner, account := range chunk.State {
		if !validator.CheckEqual(owner, account) {
			return false
		}
	}

	return true
}

// SuggestChunk is called when the chain processing system thinks we should
// nominate the value for the next chunk.
// Returns nil if we don't have enough data to suggest anything
func (q *TransactionQueue) SuggestChunk() *LedgerChunk {
	transactions := []*SignedTransaction{}
	validator := q.accounts.CowCopy()
	state := make(map[string]*Account)
	for _, t := range q.Values() {
		if validator.Process(t.Transaction) {
			transactions = append(transactions, t)
		} else {
			q.Remove(t)
		}
		state[t.From] = validator.Get(t.From)
		state[t.To] = validator.Get(t.To)
		if len(transactions) == MaxChunkSize {
			break
		}
	}
	return &LedgerChunk{
		Transactions: transactions,
		State: state,
	}
}
