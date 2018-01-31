package util

import (
	"log"
)

func Shorten(name string) string {
	length := len(name)
	if length > 6 {
		length = 6
	}
	return name[:length]
}

// Send logging through here so that it's easier to manage
func Logf(tag string, publicKey string, format string, a ...interface{}) {
	log.Printf(Shorten(publicKey) + " " + format, a...)
}
