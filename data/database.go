package data

import (
	"github.com/go-pg/pg"
)

// A Database encapsulates a connection to a Postgres database.
type Database struct {
	*pg.DB
}

// Creates a new database handle designed to be used for unit tests.
func NewTestDatabase() *Database {
	return &Database{
		DB: pg.Connect(&pg.Options{
			User:     "postgres",
			Password: "test",
			Database: "test",
		}),
	}
}
