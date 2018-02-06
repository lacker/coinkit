package currency

import (
	"log"

	"github.com/emirpasic/gods/sets/treeset"

	"coinkit/consensus"
	"coinkit/util"
)

// QueueLimit defines how many items will be held in the queue at a time
const QueueLimit = 1000

// TransactionQueue keeps the transactions that are pending but have neither
// been rejected nor confirmed.
// TransactionQueue is not threadsafe.
type TransactionQueue struct {
	// Just for logging
	publicKey string

	// The pool of pending transactions.
	set *treeset.Set

	// The ledger chunks that are being considered
	// They are indexed by their hash
	chunks map[consensus.SlotValue]*LedgerChunk

	// Ledger chunks that already got finalized
	// They are indexed by slot
	oldChunks map[int]*LedgerChunk

	// accounts is used to validate transactions
	// For now this is the actual authentic store of account data
	// TODO: get this into a real database
	accounts *AccountMap

	// The key of the last chunk to get finalized
	last consensus.SlotValue

	// The current slot we are working on
	slot int

	// A count of the number of transactions this queue has finalized
	finalized int
}

func NewTransactionQueue(publicKey string) *TransactionQueue {
	return &TransactionQueue{
		publicKey: publicKey,
		set:       treeset.NewWith(HighestPriorityFirst),
		chunks:    make(map[consensus.SlotValue]*LedgerChunk),
		oldChunks: make(map[int]*LedgerChunk),
		accounts:  NewAccountMap(),
		last:      consensus.SlotValue(""),
		slot:      1,
		finalized: 0,
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

func (q *TransactionQueue) Logf(format string, a ...interface{}) {
	util.Logf("TQ", q.publicKey, format, a...)
}

// Add adds a transaction to the queue
// If it isn't valid, we just discard it.
// We don't constantly revalidate so it's possible we have invalid
// transactions in the queue.
// Returns whether any changes were made.
func (q *TransactionQueue) Add(t *SignedTransaction) bool {
	if !q.Validate(t) || q.Contains(t) {
		return false
	}

	q.Logf("saw a new transaction: %s", t.Transaction)
	q.set.Add(t)

	if q.set.Size() > QueueLimit {
		it := q.set.Iterator()
		if !it.Last() {
			log.Fatal("logical failure with treeset")
		}
		worst := it.Value()
		q.set.Remove(worst)
	}

	return q.Contains(t)
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

// SharingMessage returns the pending transactions we want to share with other nodes.
func (q *TransactionQueue) SharingMessage() *TransactionMessage {
	ts := q.Transactions()
	if len(ts) == 0 && len(q.chunks) == 0 {
		return nil
	}
	return &TransactionMessage{
		Transactions: ts,
		Chunks:       q.chunks,
	}
}

// MaxBalance is used for testing
func (q *TransactionQueue) MaxBalance() uint64 {
	return q.accounts.MaxBalance()
}

// SetBalance is used for testing
func (q *TransactionQueue) SetBalance(owner string, balance uint64) {
	q.accounts.SetBalance(owner, balance)
}

func (q *TransactionQueue) CatchupMessage(slot int) *TransactionMessage {
	chunk, ok := q.oldChunks[slot]
	if !ok {
		return nil
	}
	chunks := make(map[consensus.SlotValue]*LedgerChunk)
	chunks[chunk.Hash()] = chunk
	return &TransactionMessage{
		Transactions: []*SignedTransaction{},
		Chunks:       chunks,
	}
}

func (q *TransactionQueue) HandleInfoMessage(m *util.InfoMessage) *AccountMessage {
	if m == nil || m.Account == "" {
		return nil
	}
	output := &AccountMessage{
		I:     q.slot,
		State: make(map[string]*Account),
	}
	output.State[m.Account] = q.accounts.Get(m.Account)
	return output
}

// Handles a transaction message from another node.
// Returns whether it made any internal updates.
func (q *TransactionQueue) HandleTransactionMessage(m *TransactionMessage) bool {
	if m == nil {
		return false
	}

	updated := false
	if m.Transactions != nil {
		for _, t := range m.Transactions {
			updated = updated || q.Add(t)
		}
	}
	if m.Chunks != nil {
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
			q.Logf("learned that %s = %s", util.Shorten(string(key)), chunk)
			q.chunks[key] = chunk
			updated = true
		}
	}
	return updated
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
		State:        state,
	}
	key := chunk.Hash()
	if _, ok := q.chunks[key]; !ok {
		// We have not already created this chunk
		q.Logf("i=%d, new chunk %s -> %s", q.slot, util.Shorten(string(key)), chunk)
		q.chunks[key] = chunk
	}
	return key, chunk
}

func (q *TransactionQueue) Combine(list []consensus.SlotValue) consensus.SlotValue {
	set := treeset.NewWith(HighestPriorityFirst)
	for _, v := range list {
		chunk := q.chunks[v]
		if chunk == nil {
			log.Fatalf("%s cannot combine unknown chunk %s", q.publicKey, v)
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

func (q *TransactionQueue) CanFinalize(v consensus.SlotValue) bool {
	_, ok := q.chunks[v]
	return ok
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

	q.oldChunks[q.slot] = chunk
	q.finalized += len(chunk.Transactions)
	q.last = v
	q.chunks = make(map[consensus.SlotValue]*LedgerChunk)
	q.slot += 1
	q.Revalidate()
}

func (q *TransactionQueue) Last() consensus.SlotValue {
	return q.last
}

// SuggestValue returns a chunk that is keyed by its hash
func (q *TransactionQueue) SuggestValue() (consensus.SlotValue, bool) {
	key, chunk := q.NewChunk(q.Transactions())
	if chunk == nil {
		q.Logf("has no suggestion")
		return consensus.SlotValue(""), false
	}
	q.Logf("i=%d, suggests %s = %s", q.slot, util.Shorten(string(key)), chunk)
	return key, true
}

func (q *TransactionQueue) ValidateValue(v consensus.SlotValue) bool {
	_, ok := q.chunks[v]
	return ok
}

func (q *TransactionQueue) Stats() {
	q.Logf("%d transactions finalized", q.finalized)
}

func (q *TransactionQueue) Log() {
	ts := q.Transactions()
	q.Logf("has %d pending transactions:", q.publicKey, len(ts))
	for _, t := range ts {
		q.Logf("%s", t.Transaction)
	}
}
