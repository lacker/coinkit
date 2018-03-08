package currency

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"

	"golang.org/x/crypto/sha3"

	"github.com/lacker/coinkit/consensus"
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

func NewEmptyChunk() *LedgerChunk {
	return &LedgerChunk{
		Transactions: []*SignedTransaction{},
		State:        make(map[string]*Account),
	}
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

func (c *LedgerChunk) String() string {
	return StringifyTransactions(c.Transactions)
}

func (c *LedgerChunk) Value() (driver.Value, error) {
	bytes, err := json.Marshal(c)
	return driver.Value(bytes), err
}

func (c *LedgerChunk) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return errors.New("expected []byte")
	}
	err := json.Unmarshal(bytes, c)
	if err != nil {
		return err
	}
	return nil
}
