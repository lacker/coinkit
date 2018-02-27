package util

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"reflect"
)

const OK = "ok"

type SignedMessage struct {
	// message is internal because it's redundant. This keeps it from being
	// passed around on the wire when it doesn't have to be.
	message Message

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
		message:       message,
		MessageString: ms,
		Signer:        kp.PublicKey().String(),
		Signature:     kp.Sign(ms),
	}
}

func (sm *SignedMessage) Message() Message {
	return sm.message
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
	if err != nil {
		return nil, err
	}

	// Keepalives don't have to be signed
	if answer.KeepAlive {
		return answer, nil
	}

	// Check the signature
	publicKey, err := ReadPublicKey(answer.Signer)
	if err != nil {
		return nil, err
	}
	if !Verify(publicKey, answer.MessageString, answer.Signature) {
		return nil, errors.New("signature failed verification")
	}

	answer.message, err = DecodeMessage(answer.MessageString)
	if err != nil {
		return nil, err
	}
	return answer, nil
}
