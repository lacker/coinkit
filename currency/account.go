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
type AccountMap map[string]Account

