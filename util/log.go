package util

import (
	"log"
)

// Send logging through here so that it's easier to manage
func Logf(tag string, publicKey string, format string, a ...interface{}) {
	shortNameLen := len(publicKey)
	if shortNameLen > 6 {
		shortNameLen = 6
	}
	shortName := publicKey[:shortNameLen]

	log.Printf(shortName + " " + format, a...)
}
