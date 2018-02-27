package util

import (
	"bytes"
	"encoding/gob"
	"log"
	"reflect"
	"strings"
)

type Message interface {
	// Slot returns 0 if the message doesn't relate to a particular slot
	Slot() int

	// MessageType returns a unique short string mapping to the type
	// TODO: drop MessageType
	MessageType() string

	// String() should return a short, human-readable string
	String() string
}

func RegisterMessageType(m Message) {
	// We want to remove the path qualifications so that encoding doesn't break on
	// refactors.
	// So fullName is something like "*util.FooMessage", and we just want "*FooMessage".
	fullName := reflect.TypeOf(m).String()
	parts := strings.Split(fullName, ".")
	gobName := parts[len(parts)-1]
	if gobName == "Message" {
		panic("gobName should not be Message")
	}
	if fullName[0] == '*' {
		gobName = "*" + gobName
	}
	gob.RegisterName(gobName, m)
}

// EncodeMessage gob-encodes a pointer to the Message interface
func EncodeMessage(m Message) []byte {
	if m == nil || reflect.ValueOf(m).IsNil() {
		panic("you should not EncodeMessage(nil)")
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(&m)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func DecodeMessage(encoded []byte) (Message, error) {
	b := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(b)
	var answer Message
	err := dec.Decode(&answer)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

// Useful for simulating a network transit
func EncodeThenDecode(message Message) Message {
	encoded := EncodeMessage(message)
	m, err := DecodeMessage(encoded)
	if err != nil {
		log.Fatal("encode-then-decode error:", err)
	}
	return m
}
