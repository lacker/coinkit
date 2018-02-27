package util

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"reflect"
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
