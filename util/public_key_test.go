package util

import (
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

func TestCheckBytes(t *testing.T) {
	var bytes [32]byte
	check := checkBytes(bytes[:])
	if check[0] != 175 || check[1] != 19 {
		t.Fatalf("bad check bytes")
	}
}
