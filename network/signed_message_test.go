package network

import (
	"log"
	"testing"

	"coinkit/consensus"
	"coinkit/util"
)

func TestSignedMessage(t *testing.T) {
	m := &consensus.NominationMessage{
		I: 1,
		Nom: []consensus.SlotValue{},
		Acc: []consensus.SlotValue{},
	}
	kp := util.NewKeyPairFromSecretPhrase("foo")
	sm := NewSignedMessage(kp, m)
	str := sm.Serialize()
	sm2, err := NewSignedMessageFromSerialized(str)
	if sm2 == nil {
		log.Print(err)
		t.Fatal("sm2 should not be nil")
	}
	if sm.signer != sm2.signer || sm.signature != sm2.signature {
		log.Printf("sm: %+v", sm)
		log.Printf("sm2: %+v", sm2)
		t.Fatal("sm should equal sm2")
	}
}
