package network

import (
	"crypto/sha512"
	"encoding/base64"
	"sort"
)

func HashString(x string) string {
	h := sha512.New()
	hashBytes := h.Sum([]byte(x))
	return base64.RawStdEncoding.EncodeToString(hashBytes)
}

// SeedSort sorts in a way that is repeatable depending on the seed string.
func SeedSort(seed string, input []string) []string {
	m := make(map[string]string)
	keys := []string{}
	for _, x := range input {
		hashed := HashString(seed + x)
		m[hashed] = x
		keys = append(keys, hashed)
	}
	sort.Strings(keys)
	answer := []string{}
	for _, key := range keys {
		answer = append(answer, m[key])
	}
	return answer
}
