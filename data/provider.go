package data

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
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

// Value and Scan let a ProviderArray map to a sql bigint[] with only ids
type ProviderArray []*Provider

func (ps ProviderArray) Value() (driver.Value, error) {
	strs := []string{}
	for _, p := range ps {
		strs = append(strs, string(p.ID))
	}
	return fmt.Sprintf("{%s}", strings.Join(strs, ",")), nil
}

func (ps *ProviderArray) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("could not stringify provider array from sql: %#v", src)
	}
	trimmed := strings.Trim(str, "{}")
	strs := strings.Split(trimmed, ",")
	answer := []*Provider{}
	for _, str := range strs {
		i, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		answer = append(answer, &Provider{
			ID: i,
		})
	}
	*ps = answer
	return nil
}
