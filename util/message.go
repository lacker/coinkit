package util

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
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

// MessageTypeMap maps into struct types whose pointer-types implement Message.
// For example, *NominationMessage is a Message. So this map contains the
// NominationMessage type.
var MessageTypeMap map[string]reflect.Type = make(map[string]reflect.Type)

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

	// TODO: drop the stuff after here
	name := m.MessageType()
	_, ok := MessageTypeMap[name]
	if ok {
		log.Fatalf("message type registered multiple times: %s", name)
	}
	mv := reflect.ValueOf(m)
	if mv.Kind() != reflect.Ptr {
		log.Fatalf("RegisterMessageType should only be called on pointers")
	}

	sv := mv.Elem()
	if sv.Kind() != reflect.Struct {
		log.Fatalf("RegisterMessageType should be called on pointers to structs")
	}

	// log.Printf("registering %s -> %+v", name, sv.Type())
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
		log.Fatal("encode-then-decode error:", err)
	}
	return m
}
