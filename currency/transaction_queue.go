package currency

import (
	"log"

	"github.com/emirpasic/gods/sets/treeset"

	"coinkit/consensus"
)

// QueueLimit defines how many items will be held in the queue at a time
const QueueLimit = 1000

// TransactionQueue keeps the transactions that are pending but have neither
// been rejected nor confirmed.
type TransactionQueue struct {
	set *treeset.Set

	// Transactions that we have not yet shared
	outbox []*SignedTransaction

	// The ledger chunks that are being considered
	// They are indexed by their hash
	chunks map[consensus.SlotValue]*LedgerChunk
	
	// accounts is used to validate transactions
	// For now this is the actual authentic store of account data
	// TODO: get this into a real database
	accounts *AccountMap

	// The key of the last chunk to get finalized
	last consensus.SlotValue
}

func NewTransactionQueue() *TransactionQueue {
	return &TransactionQueue{
		set: treeset.NewWith(HighestPriorityFirst),
		outbox: []*SignedTransaction{},
		chunks: make(map[consensus.SlotValue]*LedgerChunk),
		accounts: NewAccountMap(),
		last: consensus.SlotValue(""),
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

func (q *TransactionQueue) Contains(t *SignedTransaction) bool {
	return q.set.Contains(t)
}

func (q *TransactionQueue) Transactions() []*SignedTransaction {
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
		Chunks: q.chunks,
	}
}

// Handles a transaction message from another node.
func (q *TransactionQueue) Handle(m *TransactionMessage) {
	if m == nil {
		return
	}
	for _, t := range m.Transactions {
		q.Add(t)
	}
	for key, chunk := range m.Chunks {
		if _, ok := q.chunks[key]; ok {
			continue
		}
		if !q.accounts.ValidateChunk(chunk) {
			continue
		}
		if chunk.Hash() != key {
			continue
		}
		q.chunks[key] = chunk
	}
}

func (q *TransactionQueue) Size() int {
	return q.set.Size()
}

func (q *TransactionQueue) Validate(t *SignedTransaction) bool {
	return t != nil && t.Verify() && q.accounts.Validate(t.Transaction)
}

// Revalidate checks all pending transactions to see if they are still valid
func (q *TransactionQueue) Revalidate() {
	for _, t := range q.Transactions() {
		if !q.Validate(t) {
			q.Remove(t)
		}
	}
}

// NewLedgerChunk creates a ledger chunk from a list of signed transactions.
// The list should already be sorted and deduped and the signed transactions
// should be verified.
// Returns "", nil if there were no valid transactions.
// This adds a cache entry to q.chunks
func (q *TransactionQueue) NewChunk(
	ts []*SignedTransaction) (consensus.SlotValue, *LedgerChunk) {
	var last *SignedTransaction
	transactions := []*SignedTransaction{}
	validator := q.accounts.CowCopy()
	state := make(map[string]*Account)
	for _, t := range ts {
		if last != nil && HighestPriorityFirst(last, t) >= 0 {
			panic("NewLedgerChunk called on non-sorted list")
		}
		last = t
		if validator.Process(t.Transaction) {
			transactions = append(transactions, t)
		}
		state[t.From] = validator.Get(t.From)
		state[t.To] = validator.Get(t.To)
		if len(transactions) == MaxChunkSize {
			break
		}
	}
	if len(transactions) == 0 {
		return consensus.SlotValue(""), nil
	}
	chunk := &LedgerChunk{
		Transactions: transactions,
		State: state,
	}
	key := chunk.Hash()
	q.chunks[key] = chunk
	return key, chunk
}

func (q *TransactionQueue) Combine(list []consensus.SlotValue) consensus.SlotValue {
	set := treeset.NewWith(HighestPriorityFirst)
	for _, v := range list {
		chunk := q.chunks[v]
		if chunk == nil {
			panic("cannot combine chunks with unknown data")
		}
		for _, t := range chunk.Transactions {
			set.Add(t)
		}
	}
	transactions := []*SignedTransaction{}
	for _, t := range set.Values() {
		transactions = append(transactions, t.(*SignedTransaction))
	}
	value, chunk := q.NewChunk(transactions)
	if chunk == nil {
		panic("combining valid chunks led to nothing")
	}
	return value
}

func (q *TransactionQueue) Finalize(v consensus.SlotValue) {
	chunk, ok := q.chunks[v]
	if !ok {
		panic("We are finalizing a chunk but we don't know its data.")
	}
	
	if !q.accounts.ValidateChunk(chunk) {
		panic("We could not validate a finalized chunk.")
	}

	if !q.accounts.ProcessChunk(chunk) {
		panic("We could not process a finalized chunk.")
	}

	q.last = v
	q.chunks = make(map[consensus.SlotValue]*LedgerChunk)
	q.Revalidate()
}

func (q *TransactionQueue) Last() consensus.SlotValue {
	return q.last
}

// SuggestValue returns a chunk that is keyed by its hash
func (q *TransactionQueue) SuggestValue() (consensus.SlotValue, bool) {
	key, chunk := q.NewChunk(q.Transactions())
	if chunk == nil {
		return consensus.SlotValue(""), false
	}
	return key, true
}
