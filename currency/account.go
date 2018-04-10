package currency

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Accounts are stored in units of nanocoins.
const OneMillion = 1000 * 1000
const NumCoins = 21 * OneMillion
const OneBillion = 1000 * OneMillion
const TotalMoney = NumCoins * OneBillion

type Account struct {
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
	return fmt.Sprintf("s%d:b%d", a.Sequence, a.Balance)
}

func (a Account) Bytes() []byte {
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.LittleEndian, a)
	return buffer.Bytes()
}
