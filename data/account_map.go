package data

import (
	"github.com/lacker/coinkit/util"
)

// Used to map a public key to its Account
type AccountMap struct {
	// Storing real account data
	data map[string]*Account

	// We use the fallback when we don't have data on an account
	// Can be nil
	fallback *AccountMap
}

func NewAccountMap() *AccountMap {
	return &AccountMap{
		data: make(map[string]*Account),
	}
}

// Returns a copy of this accountmap that does copy-on-write, so changes
// made won't be visible in the original
func (m *AccountMap) CowCopy() *AccountMap {
	return &AccountMap{
		data:     make(map[string]*Account),
		fallback: m,
	}
}

func (m *AccountMap) MaxBalance() uint64 {
	answer := uint64(0)
	for _, account := range m.data {
		if account.Balance > answer {
			answer = account.Balance
		}
	}
	if m.fallback != nil {
		b := m.fallback.MaxBalance()
		if b > answer {
			answer = b
		}
	}
	return answer
}

// Checks that the data in the account map is what we expect
func (m *AccountMap) CheckEqual(key string, account *Account) bool {
	a := m.Get(key)
	if a == nil && account == nil {
		return true
	}
	if a == nil || account == nil {
		return false
	}
	return a.Sequence == account.Sequence && a.Balance == account.Balance
}

func (m *AccountMap) Get(key string) *Account {
	answer := m.data[key]
	if answer == nil && m.fallback != nil {
		return m.fallback.Get(key)
	}
	return answer
}

func (m *AccountMap) Set(key string, account *Account) {
	m.data[key] = account
}

// Validate returns whether this operation is valid
func (m *AccountMap) Validate(op util.Operation) bool {
	t, ok := op.(*SendOperation)
	if !ok {
		panic("AccountMap cannot validate non-SendOperation operations")
	}
	account := m.Get(t.Signer)
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

func (m *AccountMap) SetBalance(owner string, amount uint64) {
	oldAccount := m.Get(owner)
	sequence := uint32(0)
	if oldAccount != nil {
		sequence = oldAccount.Sequence
	}
	m.Set(owner, &Account{Sequence: sequence, Balance: amount})
}

// Process returns false if the operation cannot be processed
func (m *AccountMap) Process(op util.Operation) bool {
	t, ok := op.(*SendOperation)
	if !ok {
		panic("AccountMap cannot process non-SendOperation operations")
	}
	if !m.Validate(t) {
		return false
	}
	source := m.Get(t.Signer)
	target := m.Get(t.To)
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
	m.Set(t.Signer, newSource)
	m.Set(t.To, newTarget)
	return true
}

// ProcessChunk returns false if the whole chunk cannot be processed.
// In this situation, the account map may be left with only some of
// the operations in the chunk processed and would in practice have to be discarded.
func (m *AccountMap) ProcessChunk(chunk *LedgerChunk) bool {
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
func (m *AccountMap) ValidateChunk(chunk *LedgerChunk) bool {
	copy := m.CowCopy()
	return copy.ProcessChunk(chunk)
}
