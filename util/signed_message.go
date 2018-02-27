package util

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
)

const OK = "ok"

type SignedMessage struct {
	Message       Message
	MessageString string
	Signer        string
	Signature     string

	// Whenever keepalive is true, the SignedMessage has no real content, it's
	// just a small value used to keep a network connection alive
	KeepAlive bool
}

func NewSignedMessage(kp *KeyPair, message Message) *SignedMessage {
	if message == nil || reflect.ValueOf(message).IsNil() {
		log.Fatal("cannot sign nil message")
	}
	ms := EncodeMessage(message)
	return &SignedMessage{
		Message:       message,
		MessageString: ms,
		Signer:        kp.PublicKey().String(),
		Signature:     kp.Sign(ms),
	}
}

func (sm *SignedMessage) Serialize() string {
	return fmt.Sprintf("e:%s:%s:%s", sm.Signer, sm.Signature, sm.MessageString)
}

func NewSignedMessageFromSerialized(serialized string) (*SignedMessage, error) {
	parts := strings.SplitN(serialized, ":", 4)
	if len(parts) != 4 {
		return nil, errors.New("could not find 4 parts")
	}
	version, signer, signature, ms := parts[0], parts[1], parts[2], parts[3]
	if version != "e" {
		return nil, errors.New("unrecognized version")
	}
	publicKey, err := ReadPublicKey(signer)
	if err != nil {
		return nil, err
	}
	if !Verify(publicKey, ms, signature) {
		return nil, errors.New("signature failed verification")
	}
	m, err := DecodeMessage(ms)
	if err != nil {
		return nil, err
	}
	return &SignedMessage{
		Message:       m,
		MessageString: ms,
		Signer:        signer,
		Signature:     signature,
	}, nil
}

func KeepAlive() *SignedMessage {
	return &SignedMessage{KeepAlive: true}
}

func (sm *SignedMessage) Write(w io.Writer) {
	enc := gob.NewEncoder(w)
	err := enc.Encode(sm)
	if err != nil {
		log.Printf("ignoring error in gob write: %+v", err)
	}
}

// The caller is responsible for setting any deadlines on the reader.
func ReadSignedMessage(r *bufio.Reader) (*SignedMessage, error) {
	dec := gob.NewDecoder(r)
	answer := &SignedMessage{}
	err := dec.Decode(answer)
	return answer, err
}
