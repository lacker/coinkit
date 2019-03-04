package data

import (
	"encoding/json"
	"testing"

	"github.com/lacker/coinkit/util"
)

func TestSignedOperation(t *testing.T) {
	kp := util.NewKeyPairFromSecretPhrase("yo")
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
	kp := util.NewKeyPairFromSecretPhrase("hi")
	op := &TestingOperation{
		Number: 9,
		Signer: kp.PublicKey().String(),
	}
	so := NewSignedOperation(op, kp)
	bytes := util.AlphabeticalJSONEncode(so)
	so2 := &SignedOperation{}
	err := json.Unmarshal(bytes, so2)
	if err != nil {
		t.Fatal(err)
	}
	if so2.Operation.(*TestingOperation).Number != 9 {
		t.Fatalf("so2.Operation is %+v", so2.Operation)
	}
}

func TestSignedOperationDecodingAlsoVerifiesSignature(t *testing.T) {
	kp := util.NewKeyPairFromSecretPhrase("borp")
	op := &TestingOperation{
		Number: 10,
		Signer: kp.PublicKey().String(),
	}
	so := NewSignedOperation(op, kp)
	so.Signature = "BadSignature"
	bytes, err := json.Marshal(so)
	if err != nil {
		t.Fatal(err)
	}
	so2 := &SignedOperation{}
	err = json.Unmarshal(bytes, so2)
	if err == nil {
		t.Fatal("expected error in decoding")
	}
}

func TestSignedOperationDecodingAlsoVerifiesOperation(t *testing.T) {
	kp := util.NewKeyPairFromSecretPhrase("bop")
	op := &TestingOperation{
		Number:  11,
		Signer:  kp.PublicKey().String(),
		Invalid: true,
	}
	so := NewSignedOperation(op, kp)
	bytes, err := json.Marshal(so)
	if err != nil {
		t.Fatal(err)
	}
	so2 := &SignedOperation{}
	err = json.Unmarshal(bytes, so2)
	if err == nil {
		t.Fatal("expected error in decoding")
	}
}

func TestSignedOperationInSignedMessage(t *testing.T) {
	s := `e:0x32652ebe42a8d56314b8b11abf51c01916a238920c1f16db597ee87374515f4609d3:dak3Jy9lAdyrpjNL3Mlzse6+/BmX6EYTiJZVG9FAnteUiG/IxS1XrnIFyKbyb+S/nygveflkOgcRjAQIdaagAw:{"type":"Operation","message":{"operations":[{"operation":{"signer":"0x32652ebe42a8d56314b8b11abf51c01916a238920c1f16db597ee87374515f4609d3","capacity":100,"fee":0,"sequence":1},"type":"CreateProvider","signature":"7oAcPtGDW6togW9yv2eBaP1IfxVnciBJoylNIB38qlVHuJXkpR6cJCALPUKKGx0/Mc0dPglIYLU+Cv9xBkthCQ"}]}}`
	sm, err := util.NewSignedMessageFromSerialized(s)
	if sm == nil {
		util.Logger.Print(err)
		t.Fatal("initial deserialization failed")
	}

	s2 := sm.Serialize()
	sm2, err := util.NewSignedMessageFromSerialized(s2)
	if sm2 == nil {
		util.Logger.Print(err)
		t.Fatal("serialize-then-deserialize failed")
	}
}
