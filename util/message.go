package util

import (
	"log"
	"reflect"
)

type Message interface {
	// Slot returns 0 if the message doesn't relate to a particular slot
	Slot() int

	// MessageType returns a unique short string mapping to the type
	MessageType() string
}

var MessageTypeMap map[string]reflect.Type = make(map[string]reflect.Type)

func RegisterMessageType(m Message) {
	name := m.MessageType()
	_, ok := MessageTypeMap[name]
	if ok {
		log.Fatalf("message type registered multiple times: %s", name)
	}
	t := reflect.TypeOf(m)
	if t.Kind() != reflect.Ptr {
		log.Fatalf("RegisterMessageType should only be called on pointers")
	}
	
	MessageTypeMap[name] = t
}

