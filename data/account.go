package data

import (
	"fmt"

	"github.com/lacker/coinkit/util"
)

// Accounts are stored in units of nanocoins.
const OneMillion = 1000 * 1000
const NumCoins = 21 * OneMillion
const OneBillion = 1000 * OneMillion
const TotalMoney = NumCoins * OneBillion

type Account struct {
	Owner string

	// The sequence id of the last operation authorized by this account.
	// 0 means there have never been any authorized operations.
	// Used to prevent replay attacks.
	Sequence uint32

	// The current balance of this account.
	Balance uint64
}

// For debugging
func StringifyAccount(a *Account) string {
	if a == nil {
		return "nil"
	}
	return fmt.Sprintf("%s:s%d:b%d", util.Shorten(a.Owner), a.Sequence, a.Balance)
}

func (a *Account) CheckEqual(other *Account) error {
	if a == nil && other == nil {
		return nil
	}
	if a == nil || other == nil {
		return fmt.Errorf("a != other. a is %+v, other is %+v", a, other)
	}
	if a.Owner != other.Owner {
		return fmt.Errorf("owner %s != owner %s", a.Owner, other.Owner)
	}
	if a.Sequence != other.Sequence {
		return fmt.Errorf("data mismatch for owner %s: seq %d != seq %d",
			a.Owner, a.Sequence, other.Sequence)
	}
	if a.Balance != other.Balance {
		return fmt.Errorf("data mismatch for owner %s: balance %d != balance %d",
			a.Owner, a.Balance, other.Balance)
	}
	return nil
}

func (a *Account) Bytes() []byte {
	return []byte(fmt.Sprintf("%s:%d:%d", a.Owner, a.Sequence, a.Balance))
}
