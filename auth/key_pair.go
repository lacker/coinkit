package auth

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/ed25519"
)

type KeyPair struct {
	publicKey ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

// Generates a key pair at random
func NewKeyPair() *KeyPair {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	return &KeyPair{publicKey: pub, privateKey: priv}
}

func NewKeyPairFromSecretPhrase(phrase string) *KeyPair {
	// ed25519 needs 32 bytes of "entropy".
	// Repeat the passphrase with commas to pad it out
	for len(phrase) < 32 {
		phrase = phrase + "," + phrase
	}
	reader := bytes.NewReader([]byte(phrase))
	pub, priv, err := ed25519.GenerateKey(reader)
	if err != nil {
		panic(err)
	}
	return &KeyPair{publicKey: pub, privateKey: priv}
}

// A transportable version of the public key, using base64
func (kp *KeyPair) PublicKey() string {
	return base64.StdEncoding.EncodeToString(kp.publicKey)
}

// Interprets the message as utf8, then returns the signature as base64.
func (kp *KeyPair) Sign(message string) string {
	signature, err := kp.privateKey.Sign(rand.Reader, []byte(message), crypto.Hash(0))
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(signature)
}

// The external versions: message is handled as utf8, the keys and sigs are base64.
func Verify(publicKey string, message string, signature string) bool {
	pub, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return false
	}
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(pub, []byte(message), sig)
}
