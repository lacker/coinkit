package auth

import (
	"testing"
)

func TestRejectingGarbage(t *testing.T) {
	if Verify("garbagekey", "message", "garbagesig") {
		t.Fatal("this should not have been verified")
	}
}

func TestNewKeyPair(t *testing.T) {
	kp := NewKeyPair()
	message1 := "This is my message. There are many like it, but this one is mine."
	sig1 := kp.Sign(message1)
	message2 := "Another message"
	sig2 := kp.Sign(message2)
	if !Verify(kp.PublicKey(), message1, sig1) {
		t.Fatal("this should verify")
	}
	if !Verify(kp.PublicKey(), message2, sig2) {
		t.Fatal("this should verify")
	}
	if Verify(kp.PublicKey(), message1, sig2) {
		t.Fatal("this should not verify")
	}
	if Verify(kp.PublicKey(), message2, sig1) {
		t.Fatal("this should not verify")
	}
}

func TestNewKeyPairFromSecretPhrase(t *testing.T) {
	// TODO
}
