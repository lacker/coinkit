package util

import (
	"bufio"
	"bytes"
	"log"
	"testing"
)

func TestSignedMessage(t *testing.T) {
	m := &TestingMessage{Number: 4}
	kp := NewKeyPairFromSecretPhrase("foo")
	sm := NewSignedMessage(kp, m)
	str := sm.Serialize()
	sm2, err := NewSignedMessageFromSerialized(str)
	if sm2 == nil {
		log.Print(err)
		t.Fatal("sm2 should not be nil")
	}
	if sm.Signer != sm2.Signer || sm.Signature != sm2.Signature {
		log.Printf("sm: %+v", sm)
		log.Printf("sm2: %+v", sm2)
		t.Fatal("sm should equal sm2")
	}
}

func TestWriteThenRead(t *testing.T) {
	ka := KeepAlive()
	var b bytes.Buffer
	ka.Write(&b)
	ka2, err := ReadSignedMessage(bufio.NewReader(&b))
	if err != nil {
		t.Fatal(err)
	}
	if !ka2.KeepAlive {
		t.Fatal("the keepalive bit got lost")
	}
}
