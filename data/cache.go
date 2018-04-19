package data

import (
	"log"
)

// The Cache stores a subset of the information that is in the database. Typically
// this is the subset needed to validate some of the pending operations, so that
// we can revalidate quickly.
// Cache is not multithreaded.
// If there are multiple Cache objects for a single Database, data can get stale, so
// don't do that.
type Cache struct {
	// Storing real account data.
	// The key of the map is the owner of that account.
	// nil means there is currently no account for that owner
	data map[string]*Account

	// When we are doing a read operation and we don't have data, we can use the
	// readOnly cache. This is useful so that we can make copy-on-write versions of
	// this data, so that we can test destructive sequences of operations without
	// modifying the original.
	// readOnly can be nil.
	// readOnly and database should not both be non-nil.
	readOnly *Cache

	// When database is non-nil, writes to the cache get written through to the
	// database, and read operations look at the database when data does not
	// have the relevant data.
	// readOnly and database should not both be non-nil.
	database *Database
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]*Account),
	}
}

func NewDatabaseCache(database *Database) *Cache {
	c := NewCache()
	c.database = database
	return c
}

// Returns a copy of this cache that does copy-on-write, so changes
// made won't be visible in the original
func (cache *Cache) CowCopy() *Cache {
	c := NewCache()
	c.readOnly = cache
	return c
}

func (c *Cache) MaxBalance() uint64 {
	if c.database != nil {
		return c.database.MaxBalance()
	}
	answer := uint64(0)
	for _, account := range c.data {
		if account.Balance > answer {
			answer = account.Balance
		}
	}
	if c.readOnly != nil {
		b := c.readOnly.MaxBalance()
		if b > answer {
			answer = b
		}
	}
	return answer
}

// Checks that the data for an account is what we expect
func (c *Cache) CheckEqual(key string, account *Account) bool {
	a := c.GetAccount(key)
	if a == nil && account == nil {
		return true
	}
	if a == nil || account == nil {
		return false
	}
	return a.Sequence == account.Sequence && a.Balance == account.Balance
}

func (c *Cache) GetAccount(owner string) *Account {
	answer, ok := c.data[owner]
	if ok {
		return answer
	}

	if c.readOnly != nil {
		// When there is no direct database, we don't need to cache reads.
		answer = c.readOnly.GetAccount(owner)
	} else if c.database != nil {
		// When there is a database, we should cache reads to reduce database access.
		answer = c.database.GetAccount(owner)
		c.data[owner] = answer
	}

	if answer != nil && answer.Owner != owner {
		log.Fatalf("tried to get account with owner %s but got %+v", owner, answer)
	}
	return answer
}

func (c *Cache) UpsertAccount(account *Account) {
	if account == nil {
		log.Fatal("cannot upsert nil account")
	}
	if account.Owner == "" {
		log.Fatal("cannot upsert with no owner")
	}
	c.data[account.Owner] = account
	if c.database != nil {
		c.database.UpsertAccount(account)
	}
}

// Validate returns whether this operation is valid
func (c *Cache) Validate(op Operation) bool {
	t, ok := op.(*SendOperation)
	if !ok {
		panic("Cache cannot validate non-SendOperation operations")
	}
	account := c.GetAccount(t.Signer)
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

func (c *Cache) SetBalance(owner string, amount uint64) {
	oldAccount := c.GetAccount(owner)
	sequence := uint32(0)
	if oldAccount != nil {
		sequence = oldAccount.Sequence
	}
	c.UpsertAccount(&Account{
		Owner:    owner,
		Sequence: sequence,
		Balance:  amount,
	})
}

// Process returns false if the operation cannot be processed
func (c *Cache) Process(op Operation) bool {
	t, ok := op.(*SendOperation)
	if !ok {
		panic("Cache cannot process non-SendOperation operations")
	}
	if !c.Validate(t) {
		return false
	}
	source := c.GetAccount(t.Signer)
	target := c.GetAccount(t.To)
	if target == nil {
		target = &Account{}
	}
	newSource := &Account{
		Owner:    t.Signer,
		Sequence: t.Sequence,
		Balance:  source.Balance - t.Amount - t.Fee,
	}
	newTarget := &Account{
		Owner:    t.To,
		Sequence: target.Sequence,
		Balance:  target.Balance + t.Amount,
	}
	c.UpsertAccount(newSource)
	c.UpsertAccount(newTarget)
	return true
}

// ProcessChunk returns false if the whole chunk cannot be processed.
// In this situation, the cache may be left with only some of
// the operations in the chunk processed and would in practice have to be discarded.
func (c *Cache) ProcessChunk(chunk *LedgerChunk) bool {
	if chunk == nil {
		return false
	}
	if len(chunk.Operations) > MaxChunkSize {
		return false
	}

	for _, op := range chunk.SendOperations() {
		if op == nil || !op.Verify() || !c.Process(op) {
			if c.database != nil {
				panic("we half-processed a chunk on the database")
			}
			return false
		}
	}

	for owner, account := range chunk.State {
		if !c.CheckEqual(owner, account) {
			if c.database != nil {
				panic("we processed a chunk, but then integrity checks failed")
			}
			return false
		}
	}

	return true
}

// ValidateChunk returns true iff ProcessChunk could succeed.
func (c *Cache) ValidateChunk(chunk *LedgerChunk) bool {
	copy := c.CowCopy()
	return copy.ProcessChunk(chunk)
}
