package currency

import (
)

type Account struct {
	// The sequence id of the last transaction authorized by this account.
	// 0 means there have never been any authorized transactions.
	// Used to prevent replay attacks.
	sequence uint32

	// The current balance of this account.
	balance uint64
}

// Used to map a public key to its Account
type AccountMap map[string]*Account

// Validate returns whether this transaction is valid
func (m AccountMap) Validate(t Transaction) bool {
	account := m[t.From]
	if account == nil {
		return false
	}
	if account.sequence + 1 != t.Sequence {
		return false
	}
	cost := t.Amount + t.Fee
	if cost > account.balance {
		return false
	}

	return true
}

func (m AccountMap) SetBalance(owner string, amount uint64) {
	account := m[owner]
	if account == nil {
		account = &Account{}
		m[owner] = account
	}
	account.balance = amount
}

// Process returns false if the transaction cannot be processed
func (m AccountMap) Process(t Transaction) bool {
	if !m.Validate(t) {
		return false
	}
	source := m[t.From]
	target := m[t.To]
	if target == nil {
		target = &Account{}
		m[t.To] = target
	}
	source.sequence = t.Sequence
	source.balance -= t.Amount
	source.balance -= t.Fee
	target.balance += t.Amount
	return true
}
