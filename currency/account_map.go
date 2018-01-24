package currency

import (
)

type Account struct {
	// The sequence id of the last transaction authorized by this account.
	// 0 means there have never been any authorized transactions.
	// Used to prevent replay attacks.
	Sequence uint32

	// The current balance of this account.
	Balance uint64
}

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
		data: make(map[string]*Account),
		fallback: m,
	}
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

// Validate returns whether this transaction is valid
func (m *AccountMap) Validate(t *Transaction) bool {
	account := m.Get(t.From)
	if account == nil {
		return false
	}
	if account.Sequence + 1 != t.Sequence {
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

// Process returns false if the transaction cannot be processed
func (m *AccountMap) Process(t *Transaction) bool {
	if !m.Validate(t) {
		return false
	}
	source := m.Get(t.From)
	target := m.Get(t.To)
	if target == nil {
		target = &Account{}
	}
	newSource := &Account{
		Sequence: t.Sequence,
		Balance: source.Balance - t.Amount - t.Fee,
	}
	newTarget := &Account{
		Sequence: target.Sequence,
		Balance: target.Balance + t.Amount,
	}
	m.Set(t.From, newSource)
	m.Set(t.To, newTarget)
	return true
}
