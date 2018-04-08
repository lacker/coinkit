package currency

import (
	"github.com/emirpasic/gods/sets/treeset"

	"github.com/lacker/coinkit/consensus"
	"github.com/lacker/coinkit/util"
)

// QueueLimit defines how many items will be held in the queue at a time
const QueueLimit = 1000

// OperationQueue keeps the transactions that are pending but have neither
// been rejected nor confirmed.
// OperationQueue is not threadsafe.
type OperationQueue struct {
	// Just for logging
	publicKey util.PublicKey

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

func NewOperationQueue(publicKey util.PublicKey) *OperationQueue {
	return &OperationQueue{
		publicKey: publicKey,
		set:       treeset.NewWith(util.HighestFeeFirst),
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
func (q *OperationQueue) Top(n int) []*util.SignedOperation {
	answer := []*util.SignedOperation{}
	for _, item := range q.set.Values() {
		answer = append(answer, item.(*util.SignedOperation))
		if len(answer) == n {
			break
		}
	}
	return answer
}

// Remove removes an operation from the queue
func (q *OperationQueue) Remove(op *util.SignedOperation) {
	if op == nil {
		return
	}
	q.set.Remove(op)
}

func (q *OperationQueue) Logf(format string, a ...interface{}) {
	util.Logf("OQ", q.publicKey.ShortName(), format, a...)
}

// Add adds an operation to the queue
// If it isn't valid, we just discard it.
// We don't constantly revalidate so it's possible we have invalid
// operations in the queue, if a higher-fee operation that conflicts with a particular
// operation is added after it is.
// Returns whether any changes were made.
func (q *OperationQueue) Add(op *util.SignedOperation) bool {
	if !q.Validate(op) || q.Contains(op) {
		return false
	}

	q.Logf("saw a new operation: %s", op.Operation)
	q.set.Add(op)

	if q.set.Size() > QueueLimit {
		it := q.set.Iterator()
		if !it.Last() {
			util.Logger.Fatal("logical failure with treeset")
		}
		worst := it.Value()
		q.set.Remove(worst)
	}

	return q.Contains(op)
}

func (q *OperationQueue) Contains(op *util.SignedOperation) bool {
	return q.set.Contains(op)
}

func (q *OperationQueue) Operations() []*util.SignedOperation {
	answer := []*util.SignedOperation{}
	for _, op := range q.set.Values() {
		answer = append(answer, op.(*util.SignedOperation))
	}
	return answer
}

// TransactionMessage returns the pending transactions we want to share with other nodes.
func (q *OperationQueue) TransactionMessage() *TransactionMessage {
	ops := q.Operations()
	if len(ops) == 0 && len(q.chunks) == 0 {
		return nil
	}
	return &TransactionMessage{
		Operations: ops,
		Chunks:     q.chunks,
	}
}

// MaxBalance is used for testing
func (q *OperationQueue) MaxBalance() uint64 {
	return q.accounts.MaxBalance()
}

// SetBalance is used for testing
func (q *OperationQueue) SetBalance(owner string, balance uint64) {
	q.accounts.SetBalance(owner, balance)
}

func (q *OperationQueue) OldChunk(slot int) *LedgerChunk {
	chunk, ok := q.oldChunks[slot]
	if !ok {
		return nil
	}
	return chunk
}

func (q *OperationQueue) OldChunkMessage(slot int) *TransactionMessage {
	chunk := q.OldChunk(slot)
	if chunk == nil {
		return nil
	}
	chunks := make(map[consensus.SlotValue]*LedgerChunk)
	chunks[chunk.Hash()] = chunk
	return &TransactionMessage{
		Operations: []*util.SignedOperation{},
		Chunks:     chunks,
	}
}

func (q *OperationQueue) HandleInfoMessage(m *util.InfoMessage) *AccountMessage {
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
func (q *OperationQueue) HandleTransactionMessage(m *TransactionMessage) bool {
	if m == nil {
		return false
	}

	updated := false
	if m.Operations != nil {
		for _, op := range m.Operations {
			updated = updated || q.Add(op)
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

func (q *OperationQueue) Size() int {
	return q.set.Size()
}

func (q *OperationQueue) Validate(op *util.SignedOperation) bool {
	return op != nil && op.Verify() && q.accounts.Validate(op.Operation)
}

// Revalidate checks all pending transactions to see if they are still valid
func (q *OperationQueue) Revalidate() {
	for _, op := range q.Operations() {
		if !q.Validate(op) {
			q.Remove(op)
		}
	}
}

// NewLedgerChunk creates a ledger chunk from a list of signed transactions.
// The list should already be sorted and deduped and the signed transactions
// should be verified.
// Returns "", nil if there were no valid transactions.
// This adds a cache entry to q.chunks
func (q *OperationQueue) NewChunk(
	ops []*util.SignedOperation) (consensus.SlotValue, *LedgerChunk) {

	var last *util.SignedOperation
	validOps := []*util.SignedOperation{}
	validator := q.accounts.CowCopy()
	state := make(map[string]*Account)
	for _, op := range ops {
		if last != nil && util.HighestFeeFirst(last, op) >= 0 {
			panic("NewLedgerChunk called on non-sorted list")
		}
		last = op
		if validator.Process(op.Operation) {
			validOps = append(validOps, op)
		}
		state[op.GetSigner()] = validator.Get(op.GetSigner())

		if t, ok := op.Operation.(*SendOperation); ok {
			state[t.To] = validator.Get(t.To)
		}

		if len(validOps) == MaxChunkSize {
			break
		}
	}
	if len(ops) == 0 {
		return consensus.SlotValue(""), nil
	}
	chunk := &LedgerChunk{
		Operations: ops,
		State:      state,
	}
	key := chunk.Hash()
	if _, ok := q.chunks[key]; !ok {
		// We have not already created this chunk
		q.Logf("i=%d, new chunk %s -> %s", q.slot, util.Shorten(string(key)), chunk)
		q.chunks[key] = chunk
	}
	return key, chunk
}

func (q *OperationQueue) Combine(list []consensus.SlotValue) consensus.SlotValue {
	set := treeset.NewWith(util.HighestFeeFirst)
	for _, v := range list {
		chunk := q.chunks[v]
		if chunk == nil {
			util.Logger.Fatalf("%s cannot combine unknown chunk %s", q.publicKey, v)
		}
		for _, op := range chunk.Operations {
			set.Add(op)
		}
	}
	ops := []*util.SignedOperation{}
	for _, op := range set.Values() {
		ops = append(ops, op.(*util.SignedOperation))
	}
	value, chunk := q.NewChunk(ops)
	if chunk == nil {
		panic("combining valid chunks led to nothing")
	}
	return value
}

func (q *OperationQueue) CanFinalize(v consensus.SlotValue) bool {
	_, ok := q.chunks[v]
	return ok
}

func (q *OperationQueue) FinalizeChunk(chunk *LedgerChunk) {
	v := chunk.Hash()
	q.chunks[v] = chunk
	q.Finalize(v)
}

func (q *OperationQueue) Finalize(v consensus.SlotValue) {
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
	q.finalized += len(chunk.Operations)
	q.last = v
	q.chunks = make(map[consensus.SlotValue]*LedgerChunk)
	q.slot += 1
	q.Revalidate()
}

func (q *OperationQueue) Last() consensus.SlotValue {
	return q.last
}

// SuggestValue returns a chunk that is keyed by its hash
func (q *OperationQueue) SuggestValue() (consensus.SlotValue, bool) {
	key, chunk := q.NewChunk(q.Operations())
	if chunk == nil {
		q.Logf("has no suggestion")
		return consensus.SlotValue(""), false
	}
	q.Logf("i=%d, suggests %s = %s", q.slot, util.Shorten(string(key)), chunk)
	return key, true
}

func (q *OperationQueue) ValidateValue(v consensus.SlotValue) bool {
	_, ok := q.chunks[v]
	return ok
}

func (q *OperationQueue) Stats() {
	q.Logf("%d transactions finalized", q.finalized)
}

func (q *OperationQueue) Log() {
	ops := q.Operations()
	q.Logf("has %d pending operations:", len(ops))
	for _, op := range ops {
		q.Logf("%s", op.Operation)
	}
}
