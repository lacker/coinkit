package util

import (
	"encoding/json"

	"testing"
)

func TestSignedOperation(t *testing.T) {
	kp := NewKeyPairFromSecretPhrase("yo")
	op := &TestingOperation{
		Number: 8,
		Signer: kp.PublicKey().String(),
	}
	so := NewSignedOperation(op, kp)
	if !so.Verify() {
		t.Fatal("so should Verify")
	}
}

func TestSignedOperationJson(t *testing.T) {
	kp := NewKeyPairFromSecretPhrase("hi")
	op := &TestingOperation{
		Number: 9,
		Signer: kp.PublicKey().String(),
	}
	so := NewSignedOperation(op, kp)
	bytes, err := json.Marshal(so)
	if err != nil {
		t.Fatal(err)
	}
	so2 := &SignedOperation{}
	err = json.Unmarshal(bytes, so2)
	if err != nil {
		t.Fatal(err)
	}
	if so2.Operation.(*TestingOperation).Number != 9 {
		t.Fatalf("so2.Operation is %+v", so2.Operation)
	}
}
