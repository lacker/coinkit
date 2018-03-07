package data

import (
	"database/sql"
	"fmt"
	"log"
	"os/user"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// A Database encapsulates a connection to a Postgres database.
type Database struct {
	postgres *sqlx.DB
}

// Creates a new database handle designed to be used for unit tests.
func NewTestDatabase(i int) *Database {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	postgres := sqlx.MustConnect(
		"postgres",
		fmt.Sprintf("user=%s dbname=test%d sslmode=disable", user.Username, i))

	db := &Database{
		postgres: postgres,
	}
	db.initialize()
	return db
}

const schema = `
CREATE TABLE IF NOT EXISTS blocks (
    slot integer,
    chunk json NOT NULL,
    c integer,
    h integer
);

CREATE UNIQUE INDEX IF NOT EXISTS block_idx ON blocks (slot);
`

const blockInsert = `
INSERT INTO blocks (slot, chunk, c, h)
VALUES (:slot, :chunk, :c, :h)
`

func isUniquenessError(e error) bool {
	return strings.Contains(e.Error(), "duplicate key value violates unique constraint")
}

// initialize makes sure the schemas are set up right and panics if not
func (db *Database) initialize() {
	db.postgres.MustExec(schema)
}

// SaveBlock returns an error if it failed because this block is already saved.
// It panics if there is a fundamental database problem.
func (db *Database) SaveBlock(b *Block) error {
	_, err := db.postgres.NamedExec(blockInsert, b)
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
			log.Fatal("missing block with slot %d", slot+1)
		}
		slot += 1
		f(b)
	}
	return slot
}

func DropTestData(i int) {
	db := NewTestDatabase(i)
	db.postgres.MustExec("DROP TABLE IF EXISTS blocks")
}
