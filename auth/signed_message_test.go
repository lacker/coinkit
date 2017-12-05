package auth

import (
	"log"
	"testing"
)

func TestSignedMessage(t *testing.T) {
	kp := NewKeyPairFromSecretPhrase("foo")
	sm := NewSignedMessage(kp, "hello world")
	str := sm.Serialize()
	sm2, _ := NewSignedMessageFromSerialized(str)
	if sm.message != sm2.message || sm.signer != sm2.signer {
		log.Printf("sm: %+v", sm)
		log.Printf("sm2: %+v", sm2)
		t.Fatal("sm should equal sm2")
	}
}
