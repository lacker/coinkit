package util

import (
	"bytes"
	"encoding/gob"
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

func (m *TestingMessage) String() string {
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

func TestDecodingInvalidMessage(t *testing.T) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	enc.Encode("this string is not a valid message")
	encoded := b.Bytes()
	m, err := DecodeMessage(encoded)
	if err == nil || m != nil {
		t.Fatal("an encoded nil message should fail to decode")
	}
}
