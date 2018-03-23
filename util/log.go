package util

import (
	"log"
	"os"
)

// This is the one default global logger.
var Logger = log.New(os.Stderr, "", log.LstdFlags)

var LogType = "default"

func Shorten(name string) string {
	length := len(name)
	if length > 6 {
		length = 6
	}
	return name[:length]
}

// Send logging through here so that it's easier to manage
func Logf(tag string, publicKey string, format string, a ...interface{}) {
	Logger.Printf(tag+" "+Shorten(publicKey)+" "+format, a...)
}
