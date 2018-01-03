package network

import (
	"sort"
)

func hashString(x string) string {
	// TODO: implement something more hashy
	return "X" + x
}

// SeedSort sorts in a way that is repeatable depending on the seed string.
func SeedSort(seed string, input []string) []string {
	m := make(map[string]string)
	keys := []string{}
	for _, x := range input {
		hashed := hashString(seed + x)
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
