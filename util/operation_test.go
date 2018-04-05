package util

import (
	"encoding/json"
	"testing"
)

type TestingOperation struct {
	Number int
}

func (op *TestingOperation) OperationType() string {
	return "Testing"
}

func (op *TestingOperation) String() string {
	return "Testing"
}

func (op *TestingOperation) Signer() string {
	return "Fake Sender"
}

func (op *TestingOperation) Verify() bool {
	return true
}

func init() {
	RegisterOperationType(&TestingOperation{})
}

func TestOperationEncoding(t *testing.T) {
	op := &TestingOperation{Number: 5}
	op2 := EncodeThenDecodeOperation(op).(*TestingOperation)
	if op2.Number != 5 {
		t.Fatalf("op2.Number turned into %d", op2.Number)
	}
}

func TestDecodingInvalidOperation(t *testing.T) {
	bytes, err := json.Marshal(DecodedOperation{
		T: "Testing",
		O: nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(bytes)
	op, err := DecodeOperation(encoded)
	if err == nil || op != nil {
		t.Fatal("an encoded nil operation should fail to decode")
	}
}
