package data

import (
	"fmt"
	"strings"
)

type BucketQuery struct {
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	Limit    int    `json:"limit"`
	Provider uint64 `json:"provider"`
}

func (q *BucketQuery) String() string {
	parts := []string{}
	if q.Name != "" {
		parts = append(parts, fmt.Sprintf("name = %s", q.Name))
	}
	if q.Owner != "" {
		parts = append(parts, fmt.Sprintf("owner = %s", q.Owner))
	}
	if len(parts) == 0 {
		return "<empty>"
	}
	return strings.Join(parts, " ")
}
