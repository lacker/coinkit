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
		strs = append(strs, fmt.Sprintf("%d", p.ID))
	}
	answer := fmt.Sprintf("{%s}", strings.Join(strs, ","))
	return answer, nil
}

func (ps *ProviderArray) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("expected bytes from sql for provider array but got: %#v", src)
	}
	str := string(bytes)
	trimmed := strings.Trim(str, "{}")
	strs := strings.Split(trimmed, ",")
	answer := []*Provider{}
	for _, str := range strs {
		if str == "" {
			continue
		}
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
