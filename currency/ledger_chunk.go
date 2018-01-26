package currency

import (
	"encoding/base64"
	"sort"

	"golang.org/x/crypto/sha3"
	
	"coinkit/consensus"
)

// MaxChunkSize defines how many items can be put in a chunk
const MaxChunkSize = 100

// A LedgerChunk is the information in one block of the blockchain.
type LedgerChunk struct {
	Transactions []*SignedTransaction

	// The state of accounts after these transactions have been processed.
	// This only includes account information for the accounts that are
	// mentioned in the transactions.
	State map[string]*Account
}

func (c *LedgerChunk) Hash() consensus.SlotValue {
	h := sha3.New512()
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
	return consensus.SlotValue(base64.RawStdEncoding.EncodeToString(h.Sum(nil)))
}

