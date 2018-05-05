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
	util.Logger.Printf("connecting to postgres with %s", info)
	if len(config.Password) > 0 {
		util.Logger.Printf("(password hidden)")
		info = fmt.Sprintf("%s password=%s", info, config.Password)
	}
	postgres := sqlx.MustConnect("postgres", info)

	if config.testOnly {
		util.Logger.Printf("clearing test-only database %s", config.Database)
		postgres.Exec("DELETE FROM blocks")
		postgres.Exec("DELETE FROM accounts")
		postgres.Exec("DELETE FROM documents")
	}

	db := &Database{
		postgres: postgres,
		name:     config.Database,
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

const schema = `
CREATE TABLE IF NOT EXISTS blocks (
    slot integer,
    chunk json NOT NULL,
    c integer,
    h integer
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
	util.Logger.Printf("initializing database %s", db.name)

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

// namedExec is a helper function to execute a write within the pending transaction.
func (db *Database) namedExec(query string, arg interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.tx == nil {
		db.tx = db.postgres.MustBegin()
	}
	_, err := db.tx.NamedExec(query, arg)
	return err
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

	return nil
}

func (db *Database) AccountDataMessage(owner string) *DataMessage {
	// Use a transaction to simultaneously fetch the last block mined and
	// the data we need to respond to the message.
	// We need "repeatable read" isolation level so that those queries reflect
	// the same snapshot of the db. See:
	// https://www.postgresql.org/docs/9.1/static/transaction-iso.html
	tx := db.postgres.MustBeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})

	account := &Account{}
	err := tx.Get(account, "SELECT * FROM accounts WHERE owner=$1", owner)
	if err == sql.ErrNoRows {
		account = nil
	} else if err != nil {
		panic(err)
	}

	block := &Block{}
	var slot int
	err = tx.Get(block, "SELECT * FROM blocks ORDER BY slot DESC LIMIT 1")
	if err == sql.ErrNoRows {
		slot = 0
	} else if err != nil {
		panic(err)
	} else {
		slot = block.Slot
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	db.reads++
	return &DataMessage{
		I:        slot,
		Accounts: map[string]*Account{owner: account},
	}
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
INSERT INTO blocks (slot, chunk, c, h)
VALUES (:slot, :chunk, :c, :h)
`

func isUniquenessError(e error) bool {
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
	err := db.namedExec(blockInsert, b)
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
			util.Logger.Fatalf("missing block with slot %d", slot+1)
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
	err := db.namedExec(accountUpsert, a)
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
// It panics if there is a fundamental database problem.
// If this returns an error, the pending transaction will be unusable.
func (db *Database) InsertDocument(d *Document) error {
	err := db.namedExec(documentInsert, d)
	if err != nil {
		if isUniquenessError(err) {
			return err
		}
		panic(err)
	}
	return nil
}

func (db *Database) GetDocuments(match map[string]interface{}, limit int) []*Document {
	bytes, err := json.Marshal(match)
	if err != nil {
		panic(err)
	}
	rows, err := db.postgres.Queryx(
		"SELECT * FROM documents WHERE data @> $1 LIMIT $2", string(bytes), limit)
	db.reads++
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
	return answer
}
