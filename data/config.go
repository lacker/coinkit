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
}

func NewTestConfig(i int) *Config {
	return &Config{
		Database: fmt.Sprintf("test%d", i),
		User:     "$USER",
	}
}

func NewLocalConfig(i int) *Config {
	return &Config{
		Database: fmt.Sprintf("local%d", i),
		User:     "$USER",
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
