package data

import (
	"database/sql"
	"fmt"
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
func NewTestDatabase() *Database {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	postgres := sqlx.MustConnect(
		"postgres", fmt.Sprintf("user=%s dbname=test sslmode=disable", user.Username))

	db := &Database{
		postgres: postgres,
	}
	db.initialize()
	return db
}

const schema = `
CREATE TABLE IF NOT EXISTS blocks (
    slot integer,
    value text,
    c integer,
    h integer
);

CREATE UNIQUE INDEX IF NOT EXISTS block_idx ON blocks (slot);
`

const blockInsert = `
INSERT INTO blocks (slot, value, c, h)
VALUES (:slot, :value, :c, :h)
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
