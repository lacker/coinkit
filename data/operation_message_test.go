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
	serialized := "e:0x5b8f312caed13ac35805c69e889d24bbd3df7d6285fbca173cce47e7402a5d0bddf3:/zpIpa4ZZ/1AVAvP7mnwlr1D+XAfYX+UNeFx+UvIlv0UTYUFXnRuTveao4ULm/O8tWrOzKLHP8BgAJEN05JUCg:{\"message\":{\"operations\":[{\"operation\":{\"data\":{\"foo\":\"bar\"},\"fee\":1,\"sequence\":1,\"signer\":\"0x5b8f312caed13ac35805c69e889d24bbd3df7d6285fbca173cce47e7402a5d0bddf3\"},\"signature\":\"powQVmQmIPLMs8InVatDw0MY3Olc4G3P8p6CE/ikgVElad6cXW0jCpFC9pD0bIOAHZmXS80U9RPKUupZSA92BQ\",\"type\":\"CreateDocument\"}]},\"type\":\"Operation\"}"
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
