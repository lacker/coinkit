package currency

import (
	"crypto/sha512"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"

	"github.com/lacker/coinkit/consensus"
	"github.com/lacker/coinkit/util"
)

// MaxChunkSize defines how many items can be put in a chunk
const MaxChunkSize = 100

// A LedgerChunk is the information in one block of the blockchain.
type LedgerChunk struct {
	Operations []*util.SignedOperation

	// The state of accounts after these transactions have been processed.
	// This only includes account information for the accounts that are
	// mentioned in the transactions.
	State map[string]*Account
}

func NewEmptyChunk() *LedgerChunk {
	return &LedgerChunk{
		Operations: []*util.SignedOperation{},
		State:      make(map[string]*Account),
	}
}

func (c *LedgerChunk) Hash() consensus.SlotValue {
	h := sha512.New512_256()
	for _, op := range c.Operations {
		h.Write([]byte(op.Signature))
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
	return util.StringifyOperations(c.Operations)
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

// Returns only the operations that are transactions
// TODO: get rid of this
func (c *LedgerChunk) Transactions() []*Transaction {
	answer := []*Transaction{}
	for _, op := range c.Operations {
		t, ok := op.Operation.(*Transaction)
		if ok {
			answer = append(answer, t)
		}
	}
	return answer
}
