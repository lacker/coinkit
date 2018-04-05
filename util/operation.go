package util

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Operation is an interface for things that can be serialized onto the blockchain.
// Logically, the blockchain can be thought of as a sequence of operations. Any
// other data on the blockchain besides the sequence of operations is just for
// efficiency.
// The implementation is pretty similar to Message; maybe they would be less
// cut-n-pastey if I could figure out how to be smarter about it.
type Operation interface {
	// OperationType() returns a unique short string mapping to the operation type
	OperationType() string

	// String() should return a short, human-readable string
	String() string

	// Signer() is the public key of the user who needs to sign this operation
	Signer() string
}

// OperationTypeMap maps into struct types whose pointer-types implement Operation.
var OperationTypeMap map[string]reflect.Type = make(map[string]reflect.Type)

func RegisterOperationType(op Operation) {
	name := op.OperationType()
	_, ok := OperationTypeMap[name]
	if ok {
		Logger.Fatalf("operation type registered multiple times: %s", name)
	}
	opv := reflect.ValueOf(op)
	if opv.Kind() != reflect.Ptr {
		Logger.Fatalf("RegisterOperationType should only be called on pointers")
	}

	sv := opv.Elem()
	if sv.Kind() != reflect.Struct {
		Logger.Fatalf("RegisterOperationType should be called on pointers to structs")
	}

	OperationTypeMap[name] = sv.Type()
}

// DecodedOperation is just used for the encoding process.
type DecodedOperation struct {
	// The type of the operation
	T string

	// The operation itself
	O Operation
}

type PartiallyDecodedOperation struct {
	T string
	O json.RawMessage
}

func EncodeOperation(op Operation) string {
	if op == nil || reflect.ValueOf(op).IsNil() {
		panic("you should not EncodeOperation(nil)")
	}
	bytes, err := json.Marshal(DecodedOperation{
		T: op.OperationType(),
		O: op,
	})
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func DecodeOperation(encoded string) (Operation, error) {
	bytes := []byte(encoded)
	var pdo PartiallyDecodedOperation
	err := json.Unmarshal(bytes, &pdo)
	if err != nil {
		return nil, err
	}

	opType, ok := OperationTypeMap[pdo.T]
	if !ok {
		return nil, fmt.Errorf("unregistered op type: %s", pdo.T)
	}
	op := reflect.New(opType).Interface().(Operation)
	err = json.Unmarshal(pdo.O, &op)
	if err != nil {
		return nil, err
	}
	if op == nil {
		return nil, fmt.Errorf("it looks like a nil operation got encoded")
	}

	return op, nil
}

// Useful for testing
func EncodeThenDecodeOperation(operation Operation) Operation {
	encoded := EncodeOperation(operation)
	op, err := DecodeOperation(encoded)
	if err != nil {
		Logger.Fatal("EncodeThenDecodeOperation error:", err)
	}
	return op
}
