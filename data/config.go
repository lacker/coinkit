package data

import (
	"encoding/json"
	"fmt"
)

// Information we need for database access
type Config struct {
	// The database name
	Database string

	// The user name. $USER gets expanded to the username.
	User string

	// The host the database is on
	Host string

	// The port the database is on
	Port int
}

func NewTestConfig(i int) *Config {
	return &Config{
		Database: fmt.Sprintf("test%d", i),
		User:     "$USER",
		Host:     "127.0.0.1",
		Port:     5432,
	}
}

func NewConfigFromSerialized(serialized []byte) *Config {
	c := &Config{}
	err := json.Unmarshal(serialized, c)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *Config) Serialize() []byte {
	bytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(bytes, '\n')
}
