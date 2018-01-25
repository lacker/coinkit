package currency

import (
	"crypto/sha512"
	"encoding/base64"
	"sort"
)

// A LedgerChunk is the information in one block of the blockchain.
type LedgerChunk struct {
	Transactions []*SignedTransaction

	// The state of accounts after these transactions have been processed.
	// This only includes account information for the accounts that are
	// mentioned in the transactions.
	State map[string]*Account
}

func (c *LedgerChunk) Hash() string {
	h := sha512.New()
	for _, t := range c.Transactions {
		h.Write([]byte(t.Signature))
	}
	keys := []string{}
	for key, _ := range c.State {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		h.Write([]byte(key))
		account := c.State[key]
		h.Write(account.Bytes())
	}
	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}
