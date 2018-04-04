package util

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// TODO: make a Transaction an Operation
// TODO: make LedgerChunk take a list of operations
// TODO: add data-specific operations
type Operation interface {
	// OperationType() returns a unique short string mapping to the operation type
	OperationType() string

	// String() should return a short, human-readable string
	String() string
}

// OperationTypeMap maps into struct types whose pointer-types implement Operation.
var OperationTypeMap map[string]reflect.Type = make(map[string]reflect.Type)

func RegisterOperationType(op Operation) {
	name := op.MessageType()
	_, ok := OperationTypeMap[name]
	if ok {
		Logger.Fatalf("operation type registered multiple times: %s", name)
	}
	opv := reflect.ValueOf(op)
	if opv.Kind() != reflect.Ptr {
		Logger.Fatalf("RegisterOperationType should only be called on pointers")
	}

	OperationTypeMap[name] = opv.Type()
}
