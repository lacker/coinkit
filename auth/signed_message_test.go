package auth

import (
	"log"
	"testing"

	"coinkit/network"
)

func TestSignedMessage(t *testing.T) {
	m := &network.UptimeMessage{Uptime: 1337}
	kp := NewKeyPairFromSecretPhrase("foo")
	sm := NewSignedMessage(kp, m)
	str := sm.Serialize()
	sm2, _ := NewSignedMessageFromSerialized(str)
	if sm.message != sm2.message || sm.signer != sm2.signer {
		log.Printf("sm: %+v", sm)
		log.Printf("sm2: %+v", sm2)
		t.Fatal("sm should equal sm2")
	}
}
