package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os/user"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/lacker/coinkit/util"
)

// A Database encapsulates a connection to a Postgres database.
// It is threadsafe.
type Database struct {
	name     string
	postgres *sqlx.DB

	// reads generally cannot be used in a threadsafe way. Just use it for testing
	reads int

	// The mutex guards the transaction in progress and the member
	// variables below this line.
	// All writes happen via this transaction.
	mutex sync.Mutex

	// tx is nil when there is no transaction in progress
	tx *sqlx.Tx

	// To be threadsafe, don't access these directly. Use CurrentSlot() instead.
	// currentSlot is the last slot that has been finalized to the database.
	currentSlot int

	// How many commits have happened in the lifetime of this db handle
	commits int

	config *Config
}

var allDatabases = []*Database{}

func NewDatabase(config *Config) *Database {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	username := strings.Replace(config.User, "$USER", user.Username, 1)
	info := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=disable statement_timeout=%d",
		config.Host, config.Port, username, config.Database, 5000)
	// util.Logger.Printf("connecting to postgres with %s", info)
	if len(config.Password) > 0 {
		util.Logger.Printf("(password hidden)")
		info = fmt.Sprintf("%s password=%s", info, config.Password)
	}
	postgres, err := sqlx.Connect("postgres", info)
	if err != nil {
		util.Logger.Printf("failed to connect to postgres with user %s, db %s",
			username, config.Database)
		panic(err)
	}

	if config.testOnly {
		// util.Logger.Printf("clearing test-only database %s", config.Database)
		postgres.Exec("DELETE FROM blocks")
		postgres.Exec("DELETE FROM accounts")
		postgres.Exec("DELETE FROM documents")
	}

	db := &Database{
		postgres: postgres,
		name:     config.Database,
		config:   config,
	}
	db.initialize()
	allDatabases = append(allDatabases, db)
	return db
}

// Creates a new database handle designed to be used for unit tests.
// Whenever this is created, any existing data in the database is deleted.
func NewTestDatabase(i int) *Database {
	return NewDatabase(NewTestConfig(i))
}

func (db *Database) Config() *Config {
	return db.config
}

const schema = `
CREATE TABLE IF NOT EXISTS blocks (
    slot integer,
    chunk json NOT NULL,
    c integer,
    h integer,
    d json NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS block_slot_idx ON blocks (slot);

CREATE TABLE IF NOT EXISTS accounts (
    owner text,
    sequence integer CHECK (sequence >= 0),
    balance bigint CHECK (balance >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS account_owner_idx ON accounts (owner);

CREATE TABLE IF NOT EXISTS documents (
    id bigint,
    data jsonb NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS document_id_idx ON documents (id);
CREATE INDEX IF NOT EXISTS document_data_idx ON documents USING gin (data jsonb_path_ops);

CREATE TABLE IF NOT EXISTS buckets (
    name text,
    owner text,
    size integer
);

CREATE UNIQUE INDEX IF NOT EXISTS bucket_name_idx ON buckets (name);
`

// Not threadsafe, caller should hold mutex or be in init
func (db *Database) updateCurrentSlot() {
	b := db.LastBlock()
	if b == nil {
		db.currentSlot = 0
	} else {
		db.currentSlot = b.Slot
	}
}

// initialize makes sure the schemas are set up right and panics if not
func (db *Database) initialize() {
	// util.Logger.Printf("initializing database %s", db.name)

	// There are some strange errors on initialization that I don't understand.
	// Just sleep a bit and retry.
	errors := 0
	for {
		_, err := db.postgres.Exec(schema)
		if err == nil {
			if errors > 0 {
				util.Logger.Printf("db init retry successful")
			}
			db.updateCurrentSlot()
			return
		}
		util.Logger.Printf("db init error: %s", err)
		errors += 1
		if errors >= 3 {
			panic("too many db errors")
		}
		time.Sleep(time.Millisecond * time.Duration(200*errors))
	}
	panic("control should not reach here")
}

// namedExecTx is a helper function to execute a write within the pending transaction.
func (db *Database) namedExecTx(query string, arg interface{}) (sql.Result, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.tx == nil {
		db.tx = db.postgres.MustBegin()
	}
	res, err := db.tx.NamedExec(query, arg)
	return res, err
}

// execTx is a helper function to execute a write within the pending transaction.
func (db *Database) execTx(query string, args ...interface{}) (sql.Result, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.tx == nil {
		db.tx = db.postgres.MustBegin()
	}
	res, err := db.tx.Exec(query, args...)
	return res, err
}

// get is a helper function to execute a Get within the pending transaction.
func (db *Database) getTx(dest interface{}, query string, args ...interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.tx == nil {
		db.tx = db.postgres.MustBegin()
	}
	return db.tx.Get(dest, query, args...)
}

func (db *Database) CurrentSlot() int {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	return db.currentSlot
}

func (db *Database) Commits() int {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	return db.commits
}

func (db *Database) TransactionInProgress() bool {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.tx != nil
}

// Commit commits the pending transaction. If there is any error, it panics.
func (db *Database) Commit() {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.tx == nil {
		return
	}
	err := db.tx.Commit()
	if err != nil {
		panic(err)
	}
	db.tx = nil
	db.commits++
	db.updateCurrentSlot()
}

func (db *Database) Rollback() {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.tx == nil {
		return
	}
	err := db.tx.Rollback()
	if err != nil {
		panic(err)
	}
	db.tx = nil
}

// Panics if a transaction was left open
func (db *Database) AssertDone() {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.tx != nil {
		panic("a transaction was left in progress")
	}
}

// Can be used for testing so that we can find who left open a transaction.
// If you suspect a test of leaving an uncommitted transaction, call this at the
// end of it.
func CheckAllDatabasesCommitted() {
	for _, db := range allDatabases {
		if db.TransactionInProgress() {
			util.Logger.Fatalf("a transaction was left open in db %s", db.name)
		}
	}
	allDatabases = []*Database{}
}

func (db *Database) TotalSizeInfo() string {
	var answer string
	err := db.postgres.Get(
		&answer,
		"SELECT pg_size_pretty(pg_database_size($1))",
		db.name)
	if err != nil {
		return err.Error()
	}
	return answer
}

func (db *Database) HandleQueryMessage(m *QueryMessage) *DataMessage {
	if m == nil {
		return nil
	}

	if m.Account != "" {
		return db.AccountDataMessage(m.Account)
	}

	if m.Block != 0 {
		return db.BlockDataMessage(m.Block)
	}

	if m.Documents != nil {
		return db.DocumentDataMessage(m.Documents)
	}

	if m.Signature != "" {
		return db.SignatureDataMessage(m.Signature)
	}

	return nil
}

// readTransaction is a helper to let you use a transaction to fetch data that reflects
// the state of the blockchain at a known slot.
// It returns a transaction you can use for reads and the slot that this transaction
// reflects.
// Be sure to call finishReadTransaction(tx) when you are done with it.
func (db *Database) readTransaction() (*sqlx.Tx, int) {
	// We need "repeatable read" isolation level so that those queries reflect
	// the same snapshot of the db. See:
	// https://www.postgresql.org/docs/9.1/static/transaction-iso.html
	tx := db.postgres.MustBeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})

	block := &Block{}
	var slot int
	err := tx.Get(block, "SELECT * FROM blocks ORDER BY slot DESC LIMIT 1")
	if err == sql.ErrNoRows {
		slot = 0
	} else if err != nil {
		panic(err)
	} else {
		slot = block.Slot
	}

	return tx, slot
}

func (db *Database) finishReadTransaction(tx *sqlx.Tx) {
	err := tx.Rollback()
	if err != nil {
		panic(err)
	}

	db.reads++
}

func (db *Database) AccountDataMessage(owner string) *DataMessage {
	tx, slot := db.readTransaction()

	account := &Account{}
	err := tx.Get(account, "SELECT * FROM accounts WHERE owner=$1", owner)
	if err == sql.ErrNoRows {
		account = nil
	} else if err != nil {
		panic(err)
	}

	db.finishReadTransaction(tx)

	return &DataMessage{
		I:        slot,
		Accounts: map[string]*Account{owner: account},
	}
}

func (db *Database) BlockDataMessage(slot int) *DataMessage {
	block := db.GetBlock(slot)
	return &DataMessage{
		Blocks: map[int]*Block{slot: block},
	}
}

func (db *Database) DocumentDataMessage(q *DocumentQuery) *DataMessage {
	docs, slot := db.GetDocuments(q.Data.content, q.Limit)
	message := &DataMessage{
		Documents: docs,
		I:         slot,
	}
	return message
}

// Currently just checks the last 20 blocks for the right operation.
// TODO: store ops by signature somewhere
func (db *Database) SignatureDataMessage(signature string) *DataMessage {
	blocks := db.TailBlocks(20)
	answer := &DataMessage{
		Operations: map[string]*SignedOperation{},
	}
	for _, block := range blocks {
		if block.Slot > answer.I {
			answer.I = block.Slot
		}
		op := block.GetOperation(signature)
		if op != nil {
			answer.Operations[signature] = op
		}
	}
	return answer
}

// CheckAccountsMatchBlocks replays the blockchain from the beginning
// and returns an error if the resulting information does not match
// the information held in the accounts.
func (db *Database) CheckAccountsMatchBlocks() error {
	cache := NewCache()
	for _, account := range Airdrop {
		cache.UpsertAccount(account)
	}
	var err error
	db.ForBlocks(func(b *Block) {
		if err == nil {
			err = cache.ProcessChunk(b.Chunk)
		}
	})
	if err != nil {
		return err
	}
	return cache.CheckAgainstDatabase(db)
}

//////////////
// Blocks
//////////////

const blockInsert = `
INSERT INTO blocks (slot, chunk, c, h, d)
VALUES (:slot, :chunk, :c, :h, :d)
`

func isUniquenessError(e error) bool {
	if e == nil {
		return false
	}
	return strings.Contains(e.Error(), "duplicate key value violates unique constraint")
}

// InsertBlock returns an error if it failed because this block is already saved.
// It panics if there is a fundamental database problem.
// It returns an error if this block is not unique.
// If this returns an error, the pending transaction will be unusable.
func (db *Database) InsertBlock(b *Block) error {
	if b == nil {
		util.Logger.Fatal("cannot insert nil block")
	}
	cur := db.CurrentSlot()
	if b.Slot != cur+1 {
		util.Logger.Fatalf("inserting block at slot %d but db has slot %d", b.Slot, cur)
	}
	_, err := db.namedExecTx(blockInsert, b)
	if err != nil {
		if isUniquenessError(err) {
			return err
		}
		panic(err)
	}
	return nil
}

// GetBlock returns nil if there is no block for the provided slot.
func (db *Database) GetBlock(slot int) *Block {
	answer := &Block{}
	err := db.postgres.Get(answer, "SELECT * FROM blocks WHERE slot=$1", slot)
	db.reads++
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		panic(err)
	}
	return answer
}

// LastBlock returns nil if the database has no blocks in it yet.
func (db *Database) LastBlock() *Block {
	answer := &Block{}
	err := db.postgres.Get(answer, "SELECT * FROM blocks ORDER BY slot DESC LIMIT 1")
	db.reads++
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		panic(err)
	}
	return answer
}

// TailBlocks returns the last n blocks, or all blocks if there are less than n.
// They are in reverse chronological order.
func (db *Database) TailBlocks(n int) []*Block {
	rows, err := db.postgres.Queryx("SELECT * FROM blocks ORDER BY slot DESC LIMIT $1", n)
	db.reads++
	if err != nil {
		panic(err)
	}
	answer := []*Block{}
	for rows.Next() {
		b := &Block{}
		err := rows.StructScan(b)
		if err != nil {
			panic(err)
		}
		answer = append(answer, b)
	}
	return answer
}

// ForBlocks calls f on each block in the db, from lowest to highest number.
// It returns the number of blocks that were processed.
func (db *Database) ForBlocks(f func(b *Block)) int {
	slot := 0
	rows, err := db.postgres.Queryx("SELECT * FROM blocks ORDER BY slot")
	db.reads++
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		b := &Block{}
		err := rows.StructScan(b)
		if err != nil {
			panic(err)
		}
		if b.Slot != slot+1 {
			util.Logger.Fatalf(
				"a block with slot %d exists, but no block has slot %d", b.Slot, slot+1)
		}
		slot += 1
		f(b)
	}
	return slot
}

//////////////
// Accounts
//////////////

const accountUpsert = `
INSERT INTO accounts (owner, sequence, balance)
VALUES (:owner, :sequence, :balance)
ON CONFLICT (owner) DO UPDATE
  SET sequence = EXCLUDED.sequence,
      balance = EXCLUDED.balance;
`

// Database.UpsertAccount will not finalize until Commit is called.
func (db *Database) UpsertAccount(a *Account) error {
	_, err := db.namedExecTx(accountUpsert, a)
	if err != nil {
		panic(err)
	}
	return nil
}

// GetAccount returns nil if there is no account for the given owner.
func (db *Database) GetAccount(owner string) *Account {
	answer := &Account{}
	err := db.postgres.Get(answer, "SELECT * FROM accounts WHERE owner=$1", owner)
	db.reads++
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		panic(err)
	}
	return answer
}

type DatabaseAccountIterator struct {
	rows *sqlx.Rows
}

func (iter *DatabaseAccountIterator) Next() *Account {
	if !iter.rows.Next() {
		return nil
	}
	a := &Account{}
	err := iter.rows.StructScan(a)
	if err != nil {
		panic(err)
	}
	return a
}

func (db *Database) IterAccounts() AccountIterator {
	rows, err := db.postgres.Queryx("SELECT * FROM accounts ORDER BY owner")
	db.reads++
	if err != nil {
		panic(err)
	}
	return &DatabaseAccountIterator{
		rows: rows,
	}
}

// ForAccounts calls f on each account in the db, in no particular order.
// It returns the number of accounts.
func (db *Database) ForAccounts(f func(a *Account)) int {
	count := 0
	iter := db.IterAccounts()
	for {
		a := iter.Next()
		if a == nil {
			return count
		}
		count += 1
		f(a)
	}
}

// MaxBalance is slow, so we just use it for testing
func (db *Database) MaxBalance() uint64 {
	max := uint64(0)
	db.ForAccounts(func(a *Account) {
		if a.Balance > max {
			max = a.Balance
		}
	})
	return max
}

//////////////
// Documents
//////////////

const documentInsert = `
INSERT INTO documents (id, data)
VALUES (:id, :data)
`

// InsertDocument returns an error if it failed because there is already a document with
// this id.
// It uses the transaction.
// It panics if there is a fundamental database problem.
// If this returns an error, the pending transaction will be unusable.
func (db *Database) InsertDocument(d *Document) error {
	_, err := db.namedExecTx(documentInsert, d)
	if err != nil {
		if isUniquenessError(err) {
			return err
		}
		panic(err)
	}
	return nil
}

func (db *Database) GetDocument(id uint64) *Document {
	rows, err := db.postgres.Queryx(
		"SELECT * FROM documents WHERE id = $1 LIMIT 1", id)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		d := &Document{}
		err := rows.StructScan(d)
		if err != nil {
			panic(err)
		}
		return d
	}
	return nil
}

const documentUpdate = `
UPDATE documents
SET data = :data
WHERE id = :id
`

const documentDelete = `
DELETE FROM documents
WHERE id = $1
`

// SetDocument changes the contents of the document to be precisely the provided data.
// It uses the transaction.
// This panics if there is a fundamental database error.
// This returns an error if there is no such document, and the pending
// transaction will still be usable.
// If there is no such document, no change is made.
func (db *Database) SetDocument(doc *Document) error {
	res, err := db.namedExecTx(documentUpdate, doc)
	if err != nil {
		panic(err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
	if count != 1 {
		return fmt.Errorf("expected 1 row affected, got %d", count)
	}
	return nil
}

// UpdateDocument updates the contents of the document, using the transaction.
// Errors when there is no such document.
// If this returns an error, the pending transaction will be unusable.
func (db *Database) UpdateDocument(id uint64, data *JSONObject) error {
	// It needs to do a read within the transaction
	doc := &Document{}
	err := db.getTx(doc, "SELECT * FROM documents WHERE id = $1 LIMIT 1", id)
	if err != nil {
		return err
	}
	doc.Data.UpdateWith(data)
	db.SetDocument(doc)
	return nil
}

// DeleteDocument deletes the document, using the transaction.
// It errors when there is no such document.
// If this returns an error, the pending transaction will be unusable.
func (db *Database) DeleteDocument(id uint64) error {
	res, err := db.execTx(documentDelete, id)
	if err != nil {
		panic(err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}
	if count != 1 {
		return fmt.Errorf("expected 1 row deleted, got %d", count)
	}
	return nil
}

// GetDocuments returns a list of matching documents, along with the slot that this
// data reflects.
func (db *Database) GetDocuments(match map[string]interface{}, limit int) ([]*Document, int) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	bytes, err := json.Marshal(match)
	if err != nil {
		panic(err)
	}

	tx, slot := db.readTransaction()

	rows, err := tx.Queryx(
		"SELECT * FROM documents WHERE data @> $1 LIMIT $2", string(bytes), limit)
	if err != nil {
		panic(err)
	}
	answer := []*Document{}
	for rows.Next() {
		d := &Document{}
		err := rows.StructScan(d)
		if err != nil {
			panic(err)
		}
		answer = append(answer, d)
	}

	db.finishReadTransaction(tx)

	return answer, slot
}

//////////////
// Buckets
//////////////

const bucketInsert = `
INSERT INTO buckets (name, owner, size)
VALUES (:name, :owner, :size)
`

// InsertBucket returns an error if it failed because there is already a bucket with
// this name.
// It uses the transaction.
// It panics if there is a fundamental database problem.
// If this returns an error, the pending transaction will be unusable.
func (db *Database) InsertBucket(b *Bucket) error {
	_, err := db.namedExecTx(bucketInsert, b)
	if err != nil {
		if isUniquenessError(err) {
			return err
		}
		panic(err)
	}
	return nil
}
