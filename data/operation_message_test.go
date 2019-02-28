package data

import (
	"testing"

	"github.com/lacker/coinkit/util"
)

func TestOperationMessages(t *testing.T) {
	kp1 := util.NewKeyPairFromSecretPhrase("key pair 1")
	kp2 := util.NewKeyPairFromSecretPhrase("key pair 2")
	t1 := &SendOperation{
		Sequence: 1,
		Amount:   100,
		Fee:      2,
		Signer:   kp1.PublicKey().String(),
		To:       kp2.PublicKey().String(),
	}
	t2 := &SendOperation{
		Sequence: 1,
		Amount:   50,
		Fee:      2,
		Signer:   kp2.PublicKey().String(),
		To:       kp1.PublicKey().String(),
	}
	s1 := NewSignedOperation(t1, kp1)
	s2 := NewSignedOperation(t2, kp2)
	message := NewOperationMessage(s1, s2)

	m := util.EncodeThenDecodeMessage(message).(*OperationMessage)
	if len(m.Operations) != 2 {
		t.Fatal("expected len m.Operations to be 2")
	}
	if !m.Operations[0].Verify() {
		t.Fatal("expected m.Operations[0].Verify()")
	}
	if !m.Operations[1].Verify() {
		t.Fatal("expected m.Operations[1].Verify()")
	}

}

// Also see tests of this string in TrustedClient.test.js
func TestCreateDocumentOperationMessageFromJS(t *testing.T) {
	serialized := "e:0x5b8f312caed13ac35805c69e889d24bbd3df7d6285fbca173cce47e7402a5d0bddf3:+oP058pwNltN2ZdFD+cuO/UT/BDSrJCPUWSjvm+JTQw11T/8FP3yyq59tVUh+eauHMmY7pqhJLv1CJ5OMHKQAw:{\"type\":\"Operation\",\"message\":{\"operations\":[{\"operation\":{\"signer\":\"0x5b8f312caed13ac35805c69e889d24bbd3df7d6285fbca173cce47e7402a5d0bddf3\",\"sequence\":1,\"fee\":1,\"data\":{\"foo\":\"bar\"}},\"type\":\"Create\",\"signature\":\"wIS9/HZQQn8exsAZT2mmhPPC95UBBSqSxFmCknymwRozxe//emT0vscf8eq55n4fZ0JO+4NiDpknlCi4UKYmDA\"}]}}"
	msg, err := util.NewSignedMessageFromSerialized(serialized)
	if err != nil {
		t.Fatalf("could not decode signed message: %s", err)
	}

	opm, ok := msg.Message().(*OperationMessage)
	if !ok {
		t.Fatalf("expected operation message but got %v", msg.Message())
	}

	if len(opm.Operations) != 1 {
		t.Fatalf("expected one operation but got %v", opm.Operations)
	}
}

// TODO: put some test in JS land for this. We got it from the CLI
func TestCreateProviderOperationMessageFromJS(t *testing.T) {
	serialized := `e:0x32652ebe42a8d56314b8b11abf51c01916a238920c1f16db597ee87374515f4609d3:qQQnKWUG4bj94qV+dvNprxiGVo4GdXo3Dg2vf0FixywljvWT2JsYI57mGLdR0miifyIDFsC8vE7paWwMherBCg:{"type":"Operation","message":{"operations":[{"operation":{"signer":"0x32652ebe42a8d56314b8b11abf51c01916a238920c1f16db597ee87374515f4609d3","capacity":"100","fee":0,"sequence":1},"type":"CreateProvider","signature":"NC70cgQ8FEtxhRjoGXoeHWg71Dkm+PBbD4sCY3HK/ch3mf/TIobNC0oOE1g1GF06ub/X8qxY9QpCaCFFXak2Cg"}]}}`
	_, err := util.NewSignedMessageFromSerialized(serialized)
	if err != nil {
		t.Fatalf("could not decode signed message: %s", err)
	}
}
