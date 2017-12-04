package auth

import (
	"errors"
	"fmt"
	"strings"
)

type SignedMessage struct {
	message string
	signer string
	signature string
}

func NewSignedMessage(kp *KeyPair, message string) *SignedMessage {
	return &SignedMessage{
		message: message,
		signer: kp.PublicKey(),
		signature: kp.Sign(message),
	}
}

func (sm *SignedMessage) Serialize() string {
	return fmt.Sprintf("e:%s:%s:%s", sm.signer, sm.signature, sm.message)
}

func NewSignedMessageFromSerialized(serialized string) (*SignedMessage, error) {
	parts := strings.SplitN(serialized, ":", 4)
	if len(parts) != 4 {
		return nil, errors.New("could not find 4 parts")
	}
	version, signer, signature, message := parts[0], parts[1], parts[2], parts[3]
	if version != "e" {
		return nil, errors.New("unrecognized version")
	}
	if !Verify(signer, message, signature) {
		return nil, errors.New("signature failed verification")
	}
	return &SignedMessage{message: message, signer: signer, signature: signature}, nil
}
