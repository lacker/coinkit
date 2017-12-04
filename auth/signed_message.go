package auth

import (
	"errors"
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

func NewSignedMessageFromSerialized(serialized string) (*SignedMessage, err error) {
	parts := string.SplitN(serialized, ":", 4)
	if len(parts) != 4 {
		return nil, errors.New("could not find 4 parts")
	}
	version, signer, signature, message := parts
	if version != "e" {
		return nil, errors.New("unrecognized version")
	}
	if !Verify(signer, message, signature) {
		return nil, errors.New("signature failed verification")
	}
	return &SignedMessage{message: message, signer: signer, signature: signature}
}
