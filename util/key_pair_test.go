package util

import (
	"bytes"
	"testing"
)

func TestRejectingGarbage(t *testing.T) {
	randomKey := NewKeyPair().PublicKey()
	if VerifySignature(randomKey, "message", "garbagesig") {
		t.Fatal("this should not have been verified")
	}
}

func TestNewKeyPair(t *testing.T) {
	kp := NewKeyPair()
	message1 := "This is my message. There are many like it, but this one is mine."
	sig1 := kp.Sign(message1)
	message2 := "Another message"
	sig2 := kp.Sign(message2)
	if !VerifySignature(kp.PublicKey(), message1, sig1) {
		t.Fatal("this should verify")
	}
	if !VerifySignature(kp.PublicKey(), message2, sig2) {
		t.Fatal("this should verify")
	}
	if VerifySignature(kp.PublicKey(), message1, sig2) {
		t.Fatal("this should not verify")
	}
	if VerifySignature(kp.PublicKey(), message2, sig1) {
		t.Fatal("this should not verify")
	}
}

func TestNewKeyPairFromSecretPhrase(t *testing.T) {
	kp1 := NewKeyPairFromSecretPhrase("monkey")
	kp2 := NewKeyPairFromSecretPhrase("monkey")
	message1 := "This is my message. There are many like it, but this one is mine."
	sig1 := kp1.Sign(message1)
	message2 := "Another message"
	sig2 := kp1.Sign(message2)
	for _, kp := range []*KeyPair{kp1, kp2} {
		if !VerifySignature(kp.PublicKey(), message1, sig1) {
			t.Fatal("this should verify")
		}
		if !VerifySignature(kp.PublicKey(), message2, sig2) {
			t.Fatal("this should verify")
		}
		if VerifySignature(kp.PublicKey(), message1, sig2) {
			t.Fatal("this should not verify")
		}
		if VerifySignature(kp.PublicKey(), message2, sig1) {
			t.Fatal("this should not verify")
		}
	}
}

func TestSerializingKeyPair(t *testing.T) {
	kp := NewKeyPairFromSecretPhrase("boopaboop")
	s := kp.Serialize()
	kp2 := NewKeyPairFromSerialized(s)
	if !kp.publicKey.Equal(kp2.publicKey) {
		t.Fatal("public keys not equal")
	}
	if bytes.Compare(kp.privateKey, kp2.privateKey) != 0 {
		t.Fatal("private keys not equal")
	}
}
