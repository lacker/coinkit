package util

import (
	"os"
	"strconv"
)

func GetTestLoopLength(short int64, long int64) int64 {
	arg, err := strconv.Atoi(os.Getenv("COINKIT_LONG_TESTS"))
	if err == nil && arg == 1 {
		return long
	} else {
		return short
	}
}
