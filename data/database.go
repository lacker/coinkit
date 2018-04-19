package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os/user"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/lacker/coinkit/util"
)

// A Database encapsulates a connection to a Postgres database.
type Database struct {
	name     string
	postgres *sqlx.DB
	reads    int
	writes   int
}

func NewDatabase(config *Config) *Database {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	username := strings.Replace(config.User, "$USER", user.Username, 1)
	info := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		config.Host, config.Port, username, config.Database)
	util.Logger.Printf("connecting to postgres with %s", info)
	if len(config.Password) > 0 {
		util.Logger.Printf("(password hidden)")
		info = fmt.Sprintf("%s password=%s", info, config.Password)
	}
	postgres := sqlx.MustConnect("postgres", info)

	if config.testOnly {
		util.Logger.Printf("clearing test-only database %s", config.Database)
		postgres.MustExec("DROP TABLE IF EXISTS blocks")
		postgres.MustExec("DROP TABLE IF EXISTS accounts")
		postgres.MustExec("DROP TABLE IF EXISTS documents")
	}

	db := &Database{
		postgres: postgres,
		name:     config.Database,
	}
	db.initialize()
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
			return
		}
		util.Logger.Printf("db init error: %s", err)
		errors += 1
		if errors >= 3 {
			panic("too many db errors")
		}
		time.Sleep(time.Millisecond * time.Duration(200*errors))
	}
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
func (db *Database) InsertBlock(b *Block) error {
	_, err := db.postgres.NamedExec(blockInsert, b)
	db.writes++
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
			util.Logger.Fatal("missing block with slot %d", slot+1)
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

func (db *Database) UpsertAccount(a *Account) error {
	_, err := db.postgres.NamedExec(accountUpsert, a)
	db.writes++
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
// TODO: implement via IterAccounts
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
func (db *Database) InsertDocument(d *Document) error {
	_, err := db.postgres.NamedExec(documentInsert, d)
	db.writes++
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
