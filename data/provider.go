package data

import (
	"fmt"
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
