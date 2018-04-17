package data

import ()

// The Cache stores a subset of the information that is in the database. Typically
// this is the subset needed to validate some of the pending operations, so that
// we can revalidate quickly.
type Cache struct {
	// Storing real account data
	data map[string]*Account

	// When we are doing a read operation and we don't have data, we can use the
	// readOnly cache. This is useful so that we can make copy-on-write versions of
	// this data, so that we can test destructive sequences of operations without
	// modifying the original.
	// readOnly can be nil.
	readOnly *Cache
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]*Account),
	}
}

// Returns a copy of this cache that does copy-on-write, so changes
// made won't be visible in the original
func (m *Cache) CowCopy() *Cache {
	return &Cache{
		data:     make(map[string]*Account),
		readOnly: m,
	}
}

func (m *Cache) MaxBalance() uint64 {
	answer := uint64(0)
	for _, account := range m.data {
		if account.Balance > answer {
			answer = account.Balance
		}
	}
	if m.readOnly != nil {
		b := m.readOnly.MaxBalance()
		if b > answer {
			answer = b
		}
	}
	return answer
}

// Checks that the data in the account map is what we expect
func (m *Cache) CheckEqual(key string, account *Account) bool {
	a := m.GetAccount(key)
	if a == nil && account == nil {
		return true
	}
	if a == nil || account == nil {
		return false
	}
	return a.Sequence == account.Sequence && a.Balance == account.Balance
}

// TODO: check the returned account has the right owner
func (m *Cache) GetAccount(owner string) *Account {
	answer := m.data[owner]
	if answer == nil && m.readOnly != nil {
		return m.readOnly.GetAccount(owner)
	}
	return answer
}

// TODO: owner should not be required
func (m *Cache) UpsertAccount(owner string, account *Account) {
	m.data[owner] = account
}

// Validate returns whether this operation is valid
func (m *Cache) Validate(op Operation) bool {
	t, ok := op.(*SendOperation)
	if !ok {
		panic("Cache cannot validate non-SendOperation operations")
	}
	account := m.GetAccount(t.Signer)
	if account == nil {
		return false
	}
	if account.Sequence+1 != t.Sequence {
		return false
	}
	cost := t.Amount + t.Fee
	if cost > account.Balance {
		return false
	}

	return true
}

func (m *Cache) SetBalance(owner string, amount uint64) {
	oldAccount := m.GetAccount(owner)
	sequence := uint32(0)
	if oldAccount != nil {
		sequence = oldAccount.Sequence
	}
	m.UpsertAccount(owner, &Account{Sequence: sequence, Balance: amount})
}

// Process returns false if the operation cannot be processed
func (m *Cache) Process(op Operation) bool {
	t, ok := op.(*SendOperation)
	if !ok {
		panic("Cache cannot process non-SendOperation operations")
	}
	if !m.Validate(t) {
		return false
	}
	source := m.GetAccount(t.Signer)
	target := m.GetAccount(t.To)
	if target == nil {
		target = &Account{}
	}
	newSource := &Account{
		Sequence: t.Sequence,
		Balance:  source.Balance - t.Amount - t.Fee,
	}
	newTarget := &Account{
		Sequence: target.Sequence,
		Balance:  target.Balance + t.Amount,
	}
	m.UpsertAccount(t.Signer, newSource)
	m.UpsertAccount(t.To, newTarget)
	return true
}

// ProcessChunk returns false if the whole chunk cannot be processed.
// In this situation, the account map may be left with only some of
// the operations in the chunk processed and would in practice have to be discarded.
func (m *Cache) ProcessChunk(chunk *LedgerChunk) bool {
	if chunk == nil {
		return false
	}
	if len(chunk.Operations) > MaxChunkSize {
		return false
	}

	for _, op := range chunk.SendOperations() {
		if op == nil || !op.Verify() || !m.Process(op) {
			return false
		}
	}

	for owner, account := range chunk.State {
		if !m.CheckEqual(owner, account) {
			return false
		}
	}

	return true
}

// ValidateChunk returns true iff ProcessChunk could succeed.
func (m *Cache) ValidateChunk(chunk *LedgerChunk) bool {
	copy := m.CowCopy()
	return copy.ProcessChunk(chunk)
}
