package util

import (
	"testing"
)

type TestingMessage struct {
	Number int
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
	m := &TestingMessage{Number: 7}
	m2 := EncodeThenDecode(m).(*TestingMessage)
	if m2.Number != 7 {
		t.Fatalf("m2.Number turned into %d", m2.Number)
	}
}
