package data

import (
	"os/user"

	"github.com/go-pg/pg"
)

// A Database encapsulates a connection to a Postgres database.
type Database struct {
	*pg.DB
}

// Creates a new database handle designed to be used for unit tests.
func NewTestDatabase() *Database {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	db := &Database{
		DB: pg.Connect(&pg.Options{
			User:     user.Username,
			Password: "",
			Database: "test",
		}),
	}
	db.initialize()
	return db
}

// initialize makes sure the schemas are set up right and panics if not
func (db *Database) initialize() {
	err := db.CreateTable(&Block{}, nil)
	if err != nil {
		panic(err)
	}
}
