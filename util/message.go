package util

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Message interface {
	// Slot returns 0 if the message doesn't relate to a particular slot
	Slot() int

	// MessageType returns a unique short string mapping to the type
	MessageType() string

	// String() should return a short, human-readable string
	String() string
}

// MessageTypeMap maps into struct types whose pointer-types implement Message.
// For example, *NominationMessage is a Message. So this map contains the
// NominationMessage type.
var MessageTypeMap map[string]reflect.Type = make(map[string]reflect.Type)

func RegisterMessageType(m Message) {
	name := m.MessageType()
	_, ok := MessageTypeMap[name]
	if ok {
		Logger.Fatalf("message type registered multiple times: %s", name)
	}
	mv := reflect.ValueOf(m)
	if mv.Kind() != reflect.Ptr {
		Logger.Fatalf("RegisterMessageType should only be called on pointers")
	}

	sv := mv.Elem()
	if sv.Kind() != reflect.Struct {
		Logger.Fatalf("RegisterMessageType should be called on pointers to structs")
	}

	// Logger.Printf("registering %s -> %+v", name, sv.Type())
	MessageTypeMap[name] = sv.Type()
}

// DecodedMessage is useful for json encoding and decoding, but not necessarily
// needed outside this file. Try using EncodeMessage and DecodeMessage directly.
type DecodedMessage struct {
	// The type of the message
	T string

	// The message itself
	M Message
}

type PartiallyDecodedMessage struct {
	T string
	M json.RawMessage
}

func EncodeMessage(m Message) string {
	if m == nil || reflect.ValueOf(m).IsNil() {
		panic("you should not EncodeMessage(nil)")
	}
	bytes, err := json.Marshal(DecodedMessage{
		T: m.MessageType(),
		M: m,
	})
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func DecodeMessage(encoded string) (Message, error) {
	bytes := []byte(encoded)
	var pdm PartiallyDecodedMessage
	err := json.Unmarshal(bytes, &pdm)
	if err != nil {
		return nil, err
	}

	messageType, ok := MessageTypeMap[pdm.T]
	if !ok {
		return nil, fmt.Errorf("unregistered message type: %s", pdm.T)
	}
	m := reflect.New(messageType).Interface().(Message)
	err = json.Unmarshal(pdm.M, &m)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("it looks like a nil got encoded")
	}

	return m, nil
}

// Useful for simulating a network transit
func EncodeThenDecode(message Message) Message {
	encoded := EncodeMessage(message)
	m, err := DecodeMessage(encoded)
	if err != nil {
		Logger.Fatal("encode-then-decode error:", err)
	}
	return m
}
