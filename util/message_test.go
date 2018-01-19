package util

import (
	"testing"
)

type TestingMessage struct {
	number int
}

func (m *TestingMessage) Slot() int {
	return 0
}

func (m *TestingMessage) MessageType() string {
	return "Testing"
}

func init() {
	RegisterMessageType(&TestingMessage{})
}

func TestMessageEncoding(t *testing.T) {

}
