package data

import (
	"fmt"
	"strings"
)

type BucketQuery struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

func (q *BucketQuery) String() string {
	parts := []string{}
	if q.Name != "" {
		parts = append(parts, fmt.Sprintf("name=%s"))
	}
	if q.Owner != "" {
		parts = append(parts, fmt.Sprintf("owner=%s"))
	}
	if parts.length == 0 {
		return "<empty>"
	}
	return string.Join(parts, " ")
}
