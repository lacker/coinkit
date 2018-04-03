package data

import (
	"database/sql"
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

	db := &Database{
		postgres: postgres,
		name:     config.Database,
	}
	db.initialize()
	return db
}

// Creates a new database handle designed to be used for unit tests.
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

CREATE UNIQUE INDEX IF NOT EXISTS block_idx ON blocks (slot);

CREATE TABLE IF NOT EXISTS documents (
    id bigint,
    data json NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS id_idx ON documents (id);
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
		if errors >= 10 {
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
			util.Logger.Fatal("missing block with slot %d", slot+1)
		}
		slot += 1
		f(b)
	}
	return slot
}

func DropTestData(i int) {
	db := NewTestDatabase(i)
	db.postgres.MustExec("DROP TABLE IF EXISTS blocks")
	db.postgres.MustExec("DROP TABLE IF EXISTS documents")
}
