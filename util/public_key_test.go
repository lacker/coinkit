package util

import (
	"crypto/sha512"

	"testing"
)

func TestInvalidKeys(t *testing.T) {
	invalid, err := ReadPublicKey("blah")
	if err == nil {
		t.Fatal("blah should fail")
	}
	_, err = ReadPublicKey("0xblahblahblah")
	if err == nil {
		t.Fatal("0xblah should fail")
	}
	_, err = ReadPublicKey("0x12345678901234567890123456789012345678901234567890123456789012345678")
	if err == nil {
		t.Fatal("checksums should bork things I made up")
	}
	if invalid.Validate() {
		t.Fatal("the zero key should not validate")
	}
}

func TestValidation(t *testing.T) {
	var bytes [32]byte
	for i := 0; i < 32; i++ {
		bytes[i] = byte(i)
	}
	pk := GeneratePublicKey(bytes[:])
	if !pk.Validate() {
		t.Fatal("newly created keys should validate ok")
	}
	s := pk.String()
	pk2, err := ReadPublicKey(s)
	if err != nil {
		t.Fatal("reading a newly-written key should work")
	}
	if !pk.Equal(pk2) || !pk2.Equal(pk) {
		t.Fatal("write-then-read should lead to equality")
	}

	pk3 := GeneratePublicKey(pk.WithoutChecksum())
	if !pk.Equal(pk3) {
		t.Fatal("WithoutChecksum should be undoable")
	}
}

// Testing that our Go libraries work like our JavaScript libraries
func TestCryptoBasics(t *testing.T) {
	h := sha512.New512_256()
	sum := h.Sum(nil)
	if sum[0] != 198 {
		t.Fatalf("first byte of hashed nothing should be 198")
	}

	h.Write([]byte("foo"))
	sum = h.Sum(nil)
	if sum[0] != 213 {
		t.Fatalf("first byte of hashed foo should be 213")
	}
}
