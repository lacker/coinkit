package data

import (
	"fmt"
	"log"
	"reflect"
	"sort"

	"github.com/lacker/coinkit/util"
)

// The Cache stores a subset of the information that is in the database. Generally
// this is the subset needed to validate some of the pending operations, so that
// we can revalidate quickly.
// Cache is not multithreaded.
// If there are multiple Cache objects in use with the same cache.database
// set, data can get stale, so don't do that.
//
// In general, some methods write through to the database if there is one, but leave
// it as a pending transaction. Copy-on-write means that when you make a copy and write
// to the copy, it won't write through.
// When there is a db write-through, this means you need to call db.Commit() afterwards, or
// else there is a transaction left on the database that will probably cause hangs during
// subsequent database operations. In the comments this is just referred to as
// "Foo writes through" - all writethroughs work this way.
type Cache struct {
	// Storing real account data.
	// The key of the map is the owner of that account.
	// nil means there is currently no account for that owner
	accounts map[string]*Account

	// Storing past blocks.
	blocks map[int]*Block

	// documents stores a subset of the documents in the database.
	// The key of the map is the document id.
	// nil means there is currently no such document.
	documents map[uint64]*Document

	// buckets stores a subset of the buckets in the database.
	// The key of the map is the bucket name.
	// The provider data is not stored on these buckets, only the IDs.
	buckets map[string]*Bucket

	// providers stores a subset of the providers in the database.
	// The key of the map is the provider id.
	providers map[uint64]*Provider

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

	NextDocumentID uint64
	NextProviderID uint64
}

func NewCache() *Cache {
	return &Cache{
		accounts:       make(map[string]*Account),
		blocks:         make(map[int]*Block),
		documents:      make(map[uint64]*Document),
		buckets:        make(map[string]*Bucket),
		providers:      make(map[uint64]*Provider),
		NextDocumentID: uint64(1),
		NextProviderID: uint64(1),
	}
}

func NewDatabaseCache(database *Database, nextDocumentID uint64, nextProviderID uint64) *Cache {
	c := NewCache()
	c.database = database
	c.NextDocumentID = nextDocumentID
	c.NextProviderID = nextProviderID
	return c
}

// Returns a copy of this cache that writes changes into the copy, so changes
// made won't be visible in the original, but lets reads fall through to the
// original.
func (cache *Cache) CowCopy() *Cache {
	c := NewCache()
	c.readOnly = cache
	c.NextDocumentID = cache.NextDocumentID
	c.NextProviderID = cache.NextProviderID
	return c
}

// CheckAgainstDatabase returns an error if any of the account data in the
// memory part of the cache does not match against the database.
// Typically this will run on startup to check integrity.
func (c *Cache) CheckAgainstDatabase(db *Database) error {
	if db.TransactionInProgress() {
		return fmt.Errorf("there is an uncommitted transaction")
	}

	// Check accounts
	for owner, cacheAccount := range c.accounts {
		dbAccount := db.GetAccount(owner)
		err := cacheAccount.CheckEqual(dbAccount)
		if err != nil {
			return err
		}
	}

	// Check buckets
	for name, cacheBucket := range c.buckets {
		dbBucket := db.GetBucket(name)
		err := cacheBucket.CheckEqual(dbBucket)
		if err != nil {
			return err
		}
	}

	// Check providers
	for id, cacheProvider := range c.providers {
		dbProvider := db.GetProvider(id)
		err := cacheProvider.CheckEqual(dbProvider)
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

////////////////////
// Account stuff
////////////////////

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

// Do not modify the Account returned from GetAccount, because it might belong to
// the readonly cache.
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

// UpsertAccount writes through.
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

// SetBalance writes through.
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

// ProcessSendOperation writes through.
// ProcessSendOperation returns false if the operation cannot be processed
func (c *Cache) ProcessSendOperation(op *SendOperation) bool {
	source := c.GetAccount(op.Signer)
	target := c.GetAccount(op.To)
	if target == nil {
		target = &Account{}
	}
	newSource := &Account{
		Owner:    op.Signer,
		Sequence: op.Sequence,
		Balance:  source.Balance - op.Amount - op.Fee,
	}
	newTarget := &Account{
		Owner:    op.To,
		Sequence: target.Sequence,
		Balance:  target.Balance + op.Amount,
	}
	c.UpsertAccount(newSource)
	c.UpsertAccount(newTarget)
	return true
}

// IncrementSequence writes through.
// Increments the sequence number for the provided op.
// The op should already have been validated.
func (c *Cache) IncrementSequence(op Operation) {
	account := c.GetAccount(op.GetSigner())
	if account.Sequence+1 != op.GetSequence() {
		panic("sequence numbers were not validated")
	}
	newAccount := &Account{
		Owner:    op.GetSigner(),
		Sequence: op.GetSequence(),
	}
	c.UpsertAccount(newAccount)
}

/////////////////////
// Document stuff
/////////////////////

// InsertDocument writes through.
// This does not update NextDocumentID.
func (c *Cache) InsertDocument(doc *Document) {
	c.documents[doc.ID] = doc
	if c.database != nil {
		check(c.database.InsertDocument(doc))
	}
}

// UpdateDocument writes through.
func (c *Cache) UpdateDocument(doc *Document) {
	c.documents[doc.ID] = doc

	if c.database != nil {
		check(c.database.UpdateDocument(doc.ID, doc.Data))
	}
}

// DeleteDocument writes through.
func (c *Cache) DeleteDocument(id uint64) {
	c.documents[id] = nil
	if c.database != nil {
		check(c.database.DeleteDocument(id))
	}
}

func (c *Cache) DocExists(id uint64) bool {
	return c.GetDocument(id) != nil
}

func (c *Cache) DocOwner(id uint64) string {
	doc := c.GetDocument(id)
	if doc == nil {
		return ""
	}
	return doc.Owner()
}

// Do not modify the Document returned from GetDocument, because it might belong to
// the readonly cache.
// Returns nil if there is no such document.
func (c *Cache) GetDocument(id uint64) *Document {
	doc, ok := c.documents[id]
	if ok {
		return doc
	}

	if c.readOnly != nil {
		return c.readOnly.GetDocument(id)
	}
	if c.database != nil {
		// When there is a database, read from the database and cache it.
		doc = c.database.GetDocument(id)
		c.documents[id] = doc
		return doc
	}

	return nil
}

///////////////////
// Bucket stuff
///////////////////

// GetBucket returns a copy so you could modify it without borking the cache.
// This includes the data from providers in the returned bucket whenever that data exists.
// Returns nil if there is no such bucket.
func (c *Cache) GetBucket(name string) *Bucket {
	cached, ok := c.buckets[name]
	if ok {
		// Inflate the provider data.
		bucket := cached.StripProviderData()
		for i, p := range bucket.Providers {
			provider := c.GetProvider(p.ID)
			if provider != nil {
				bucket.Providers[i] = provider
			}
		}
		return bucket
	}

	if c.readOnly != nil {
		return c.readOnly.GetBucket(name)
	}
	if c.database != nil {
		// When there is a database, read from the database and cache it.
		bucket := c.database.GetBucket(name)
		for _, p := range bucket.Providers {
			c.providers[p.ID] = p
		}
		c.buckets[name] = bucket.StripProviderData()
	}

	return nil
}

func (c *Cache) BucketExists(name string) bool {
	return c.GetBucket(name) != nil
}

func (c *Cache) BucketOwner(name string) string {
	b := c.GetBucket(name)
	if b == nil {
		return ""
	}
	return b.Owner
}

// InsertBucket writes through.
func (c *Cache) InsertBucket(b *Bucket) {
	if !b.IsValidNewBucket() {
		util.Logger.Fatalf("cannot InsertBucket: %+v", b)
	}
	c.buckets[b.Name] = b

	if c.database != nil {
		check(c.database.InsertBucket(b))
	}
}

// DeleteBucket deallocates this bucket from all providers and then deletes the bucket object.
// DeleteBucket writes through.
func (c *Cache) DeleteBucket(name string) {
	b := c.GetBucket(name)
	if b == nil {
		util.Logger.Fatalf("cannot delete nonexistent bucket: %s", name)
	}

	for _, p := range b.Providers {
		c.Deallocate(name, p.ID)
	}

	c.buckets[name] = nil
	if c.database != nil {
		check(c.database.DeleteBucket(name))
	}
}

/////////////////////
// Provider stuff
/////////////////////

// Do not modify the Provider returned from GetProvider, because it might belong to the
// readonly cache.
// Returns nil if there is no such provider.
func (c *Cache) GetProvider(id uint64) *Provider {
	p, ok := c.providers[id]
	if ok {
		return p
	}

	if c.readOnly != nil {
		return c.readOnly.GetProvider(id)
	}
	if c.database != nil {
		// When there is a database, read from the database and cache it.
		p = c.database.GetProvider(id)
		c.providers[id] = p
		return p
	}

	return nil
}

func (c *Cache) ProviderExists(id uint64) bool {
	return c.GetProvider(id) != nil
}

func (c *Cache) ProviderOwner(id uint64) string {
	p := c.GetProvider(id)
	if p == nil {
		return ""
	}
	return p.Owner
}

// InsertProvider writes through.
// This does not update NextProviderID.
func (c *Cache) InsertProvider(p *Provider) {
	if !p.IsValidNewProvider() {
		util.Logger.Fatalf("invalid new provider to insert: #%v", p)
	}
	c.providers[p.ID] = p

	if c.database != nil {
		check(c.database.InsertProvider(p))
	}
}

// AddCapacity increases the capacity on a provider.
// AddCapacity writes through.
func (c *Cache) AddCapacity(id uint64, amount uint32) {
	p := c.GetProvider(id)
	if p == nil {
		panic("no provider found for update")
	}
	c.providers[id].Capacity += amount
	c.providers[id].Available += amount

	if c.database != nil {
		err := c.database.AddCapacity(id, amount)
		check(err)
	}
}

// DeleteProvider deallocates all buckets from this provider and then deletes the
// provider object.
// DeleteProvider writes through.
func (c *Cache) DeleteProvider(id uint64) {
	p := c.GetProvider(id)
	if p == nil {
		util.Logger.Fatalf("cannot delete nonexistent provider: %d", id)
	}
	for _, b := range p.Buckets {
		c.Deallocate(b.Name, id)
	}

	c.providers[id] = nil
	if c.database != nil {
		check(c.database.DeleteProvider(id))
	}
}

/////////////////////
// Allocation stuff
/////////////////////

// Allocate writes through.
func (c *Cache) Allocate(bucketName string, providerID uint64) {
	// Check that the provider has enough space for the bucket
	b := c.GetBucket(bucketName)
	p := c.GetProvider(providerID)
	if b == nil || p == nil || b.Size > p.Available {
		panic("invalid allocation")
	}

	cachedBucket := c.buckets[bucketName]
	cachedBucket.Providers = append(cachedBucket.Providers, &Provider{ID: providerID})
	cachedProvider := c.providers[providerID]
	cachedProvider.Buckets = append(cachedProvider.Buckets, &Bucket{Name: bucketName})
	cachedProvider.Available -= b.Size

	if c.database != nil {
		check(c.database.Allocate(bucketName, providerID))
	}
}

// Deallocate writes through.
func (c *Cache) Deallocate(bucketName string, providerID uint64) {
	// Check that this pair is actually an allocation
	b := c.GetBucket(bucketName)
	p := c.GetProvider(providerID)
	if b == nil || p == nil || !b.HasProvider(providerID) || !p.HasBucket(bucketName) {
		panic("invalid deallocation")
	}

	c.buckets[bucketName].RemoveProvider(providerID)
	cachedProvider := c.providers[providerID]
	cachedProvider.RemoveBucket(bucketName)
	cachedProvider.Available += b.Size

	if c.database != nil {
		check(c.database.Deallocate(bucketName, providerID))
	}
}

/////////////////////////////////////
// General block processing stuff
/////////////////////////////////////

// Validate returns whether this operation is valid.
// Issues that can be caught just by looking at the operation itself should be checked in
// the Verify method of the operation. Validation here is for checking whether operations
// are consistent with the data that is already in the database.
func (c *Cache) Validate(operation Operation) bool {
	account := c.GetAccount(operation.GetSigner())
	if account == nil {
		return false
	}
	if account.Sequence+1 != operation.GetSequence() {
		return false
	}
	if account.Balance < operation.GetFee() {
		return false
	}

	switch op := operation.(type) {

	case *SendOperation:
		return account.ValidateSendOperation(op)

	case *CreateDocumentOperation:
		return true

	case *UpdateDocumentOperation:
		return c.DocOwner(op.ID) == op.Signer

	case *DeleteDocumentOperation:
		return c.DocOwner(op.ID) == op.Signer

	case *CreateBucketOperation:
		return !c.BucketExists(op.Name)

	case *DeleteBucketOperation:
		return c.BucketOwner(op.Name) == op.Signer

	case *CreateProviderOperation:
		return true

	case *DeleteProviderOperation:
		return c.ProviderOwner(op.ID) == op.Signer

	case *AllocateOperation:
		p := c.GetProvider(op.ID)
		b := c.GetBucket(op.Name)
		if p == nil || b == nil {
			return false
		}
		if p.Owner != op.Signer && b.Owner != op.Signer {
			return false
		}
		if p.Available < b.Size {
			return false
		}
		if p.HasBucket(op.Name) || b.HasProvider(op.ID) {
			return false
		}
		return true

	case *DeallocateOperation:
		p := c.GetProvider(op.ID)
		b := c.GetBucket(op.Name)
		if p == nil || b == nil {
			return false
		}
		if p.Owner != op.Signer && b.Owner != op.Signer {
			return false
		}
		if !p.HasBucket(op.Name) || !b.HasProvider(op.ID) {
			return false
		}
		return true

	default:
		util.Printf("operation: %+v has type %s", operation, reflect.TypeOf(operation))
		panic("operation type cannot be validated")
	}
}

// Process returns false if the operation cannot be processed
func (c *Cache) Process(operation Operation) bool {
	if !c.Validate(operation) {
		return false
	}

	switch op := operation.(type) {

	case *SendOperation:
		return c.ProcessSendOperation(op)

	case *CreateDocumentOperation:
		c.IncrementSequence(op)
		doc := NewDocumentFromOperation(c.NextDocumentID, op)
		c.InsertDocument(doc)
		c.NextDocumentID++
		return true

	case *UpdateDocumentOperation:
		c.IncrementSequence(op)
		doc := NewDocumentFromOperation(op.ID, op)
		c.UpdateDocument(doc)
		return true

	case *DeleteDocumentOperation:
		c.IncrementSequence(op)
		c.DeleteDocument(op.ID)
		return true

	case *CreateBucketOperation:
		c.IncrementSequence(op)
		bucket := &Bucket{
			Name:  op.Name,
			Owner: op.Signer,
			Size:  op.Size,
		}
		c.InsertBucket(bucket)
		return true

	case *DeleteBucketOperation:
		c.IncrementSequence(op)
		c.DeleteBucket(op.Name)
		return true

	case *CreateProviderOperation:
		c.IncrementSequence(op)
		p := &Provider{
			ID:        c.NextProviderID,
			Owner:     op.Signer,
			Capacity:  op.Capacity,
			Available: op.Capacity,
		}
		c.InsertProvider(p)
		c.NextProviderID++
		return true

	case *DeleteProviderOperation:
		c.IncrementSequence(op)
		c.DeleteProvider(op.ID)
		return true

	case *AllocateOperation:
		c.IncrementSequence(op)
		c.Allocate(op.Name, op.ID)
		return true

	default:
		util.Fatalf("unhanded type in cache.Process: %s", reflect.TypeOf(operation))
		return false
	}
	panic("you forgot to add a return statement in the cache.Process switch")
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
		check(c.database.InsertBlock(block))
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

	for _, op := range chunk.Operations {
		if op == nil {
			return fmt.Errorf("chunk has a nil op")
		}
		if !op.Verify() {
			return fmt.Errorf("op failed verify: %+v", op)
		}
		if !c.Process(op.Operation) {
			return fmt.Errorf("op failed to process: %+v", op)
		}
	}

	for owner, account := range chunk.Accounts {
		if !c.CheckEqual(owner, account) {
			return fmt.Errorf("integrity checks failed after chunk processing")
		}
	}

	if c.NextDocumentID != chunk.NextDocumentID {
		return fmt.Errorf("bad NextDocumentID")
	}

	if c.NextProviderID != chunk.NextProviderID {
		return fmt.Errorf("bad NextProviderID")
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
