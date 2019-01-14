package data

import (
	"fmt"
	"strings"
)

type ProviderQuery struct {
	ID    string `json:"string"`
	Owner string `json:"owner"`
	Limit int    `json:"limit"`
}

func (q *ProviderQuery) String() string {
	parts := []string{}
	if q.ID != "" {
		parts = append(parts, fmt.Sprintf("id=%s", q.ID))
	}
	if q.Owner != "" {
		parts = append(parts, fmt.Sprintf("owner=%s", q.Owner))
	}
	if len(parts) == 0 {
		return "<empty>"
	}
	return strings.Join(parts, " ")
}
