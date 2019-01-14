package data

import (
	"database/sql/driver"
	"fmt"

	"github.com/lacker/coinkit/util"
)

type Provider struct {
	// Every provider gets a unique id, assigned by the blockchain.
	ID uint64 `json:"id"`

	Owner string `json:"owner"`

	// Measured in megabytes
	Capacity uint32 `json:"capacity"`
}

func (p *Provider) String() string {
	return fmt.Sprintf("provider #%d, owner:%s, capacity:%d", p.ID, p.Owner, p.Capacity)
}

// Value and Scan let a ProviderArray map to a sql bigint[]
type ProviderArray []*Provider

func (ps ProviderArray) Value() (driver.Value, error) {
	util.Logger.Fatalf("TODO: implement Value()")
	return nil, nil
}

func (ps ProviderArray) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	util.Logger.Fatalf("TODO: implement Scan. Scan called on %#v", src)
	return nil
}
