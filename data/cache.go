package data

import (
	"fmt"
	"log"
	"sort"

	"github.com/lacker/coinkit/util"
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
	accounts map[string]*Account

	// Storing past blocks.
	blocks map[int]*Block

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
		accounts: make(map[string]*Account),
		blocks:   make(map[int]*Block),
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
	for _, account := range c.accounts {
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

// CheckAgainstDatabase returns an error if any of the data in the
// memory part of the cache does not match against the database.
func (c *Cache) CheckAgainstDatabase(db *Database) error {
	if db.TransactionInProgress() {
		return fmt.Errorf("there is an uncommitted transaction")
	}
	for owner, dataAccount := range c.accounts {
		dbAccount := db.GetAccount(owner)
		err := dataAccount.CheckEqual(dbAccount)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckConsistency returns an error if there is any mismatch between
// this cache and its own database.
func (c *Cache) CheckConsistency() error {
	if c.database == nil {
		return nil
	}
	return c.CheckAgainstDatabase(c.database)
}

func (c *Cache) GetAccount(owner string) *Account {
	answer, ok := c.accounts[owner]
	if ok {
		return answer
	}

	if c.readOnly != nil {
		// When there is no direct database, we don't need to cache reads.
		answer = c.readOnly.GetAccount(owner)
	} else if c.database != nil {
		// When there is a database, we should cache reads to reduce database access.
		answer = c.database.GetAccount(owner)
		c.accounts[owner] = answer
	}

	if answer != nil && answer.Owner != owner {
		log.Fatalf("tried to get account with owner %s but got %+v", owner, answer)
	}
	return answer
}

// UpsertAccount writes through to the underlying database, but it leaves it as
// a pending transaction. The caller must call db.Commit() themselves.
func (c *Cache) UpsertAccount(account *Account) {
	if account == nil {
		log.Fatal("cannot upsert nil account")
	}
	if account.Owner == "" {
		log.Fatal("cannot upsert with no owner")
	}
	c.accounts[account.Owner] = account
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

// FinalizeBlock should be called whenever a new block is mined.
// This updates account data as well as block data.
// The modification of database state happens in a single transaction so that
// other code using the database will see consistent state.
func (c *Cache) FinalizeBlock(block *Block) {
	if block.D.Threshold == 0 {
		util.Logger.Fatalf("cannot finalize with bad quorum slice: %+v", block.D)
	}

	if err := c.ValidateChunk(block.Chunk); err != nil {
		util.Logger.Fatalf("We could not validate a finalized chunk: %s", err)
	}

	if err := c.ProcessChunk(block.Chunk); err != nil {
		util.Logger.Fatalf("Failure while processing a finalized chunk: %s", err)
	}

	c.blocks[block.Slot] = block

	if c.database != nil {
		err := c.database.InsertBlock(block)
		if err != nil {
			panic(err)
		}
		c.database.Commit()
	}
}

// ProcessChunk returns an error if the whole chunk cannot be processed.
// In this situation, the cache may be left with only some of
// the operations in the chunk processed and would in practice have to be discarded.
// If this cache has a database, it is left with a transaction in progress; the
// caller of ProcessChunk must call db.Commit() themselves.
func (c *Cache) ProcessChunk(chunk *LedgerChunk) error {
	if chunk == nil {
		return fmt.Errorf("cannot process nil chunk")
	}
	if len(chunk.Operations) > MaxChunkSize {
		return fmt.Errorf("%d ops in a chunk is too many", len(chunk.Operations))
	}

	for _, op := range chunk.SendOperations() {
		if op == nil {
			return fmt.Errorf("chunk has a nil op")
		}
		if !op.Verify() {
			return fmt.Errorf("op failed verify: %+v", op)
		}
		if !c.Process(op) {
			return fmt.Errorf("op failed to process: %+v", op)
		}
	}

	for owner, account := range chunk.Accounts {
		if !c.CheckEqual(owner, account) {
			return fmt.Errorf("integrity checks failed after chunk processing")
		}
	}

	return nil
}

// ValidateChunk returns an error iff ProcessChunk would fail.
func (c *Cache) ValidateChunk(chunk *LedgerChunk) error {
	copy := c.CowCopy()
	return copy.ProcessChunk(chunk)
}

type CacheAccountIterator struct {
	nextIndex int
	accounts  []*Account
}

func (iter *CacheAccountIterator) Next() *Account {
	if iter.nextIndex >= len(iter.accounts) {
		return nil
	}
	answer := iter.accounts[iter.nextIndex]
	iter.nextIndex++
	return answer
}

func (c *Cache) IterAccounts() AccountIterator {
	if c.database != nil {
		return c.database.IterAccounts()
	}
	if c.readOnly != nil {
		panic("IterAccounts for cow copies not implemented")
	}

	// Make sure to go through owners in sorted order
	owners := []string{}
	for owner, _ := range c.accounts {
		owners = append(owners, owner)
	}
	sort.Strings(owners)

	iter := &CacheAccountIterator{
		nextIndex: 0,
		accounts:  []*Account{},
	}
	for _, owner := range owners {
		iter.accounts = append(iter.accounts, c.GetAccount(owner))
	}
	return iter
}

// GetBlock returns nil if there is no block for the provided slot.
func (c *Cache) GetBlock(slot int) *Block {
	block, ok := c.blocks[slot]
	if ok {
		return block
	}
	if c.database != nil {
		b := c.database.GetBlock(slot)
		if b.D == nil {
			util.Logger.Fatalf("database block for slot %d has nil quorum slice", slot)
		}
		return b
	}
	return nil
}
