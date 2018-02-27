package util

import (
	"bufio"
	"bytes"
	"log"
	"testing"
)

func TestKeepAlive(t *testing.T) {
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

func TestNormalSignedMessage(t *testing.T) {
	m := &TestingMessage{Number: 4}
	kp := NewKeyPairFromSecretPhrase("foo")
	sm := NewSignedMessage(kp, m)

	var b bytes.Buffer
	sm.Write(&b)
	sm2, err := ReadSignedMessage(bufio.NewReader(&b))

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

func TestCantSwapSignature(t *testing.T) {
	m1 := &TestingMessage{Number: 5}
	m2 := &TestingMessage{Number: 6}
	kp := NewKeyPairFromSecretPhrase("bork")
	sm1 := NewSignedMessage(kp, m1)
	sm2 := NewSignedMessage(kp, m2)

	// Corrupt sm1
	sm1.Signature = sm2.Signature

	var b bytes.Buffer
	sm1.Write(&b)
	out, err := ReadSignedMessage(bufio.NewReader(&b))

	if out != nil || err == nil {
		t.Fatal("this should be caught as a bad signature")
	}
}
