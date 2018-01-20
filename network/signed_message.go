package network

import (
	"errors"
	"fmt"
	"strings"

	"coinkit/util"
)

type SignedMessage struct {
	message util.Message
	messageString string
	signer string
	signature string
}

func NewSignedMessage(kp *util.KeyPair, message util.Message) *SignedMessage {
	ms := util.EncodeMessage(message)
	return &SignedMessage{
		message: message,
		messageString: ms,
		signer: kp.PublicKey(),
		signature: kp.Sign(ms),
	}
}

func (sm *SignedMessage) Message() util.Message {
	return sm.message
}

func (sm *SignedMessage) Signer() string {
	return sm.signer
}

func (sm *SignedMessage) Serialize() string {
	return fmt.Sprintf("e:%s:%s:%s", sm.signer, sm.signature, sm.messageString)
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
	if !util.Verify(signer, ms, signature) {
		return nil, errors.New("signature failed verification")
	}
	m, err := util.DecodeMessage(ms)
	if err != nil {
		return nil, err
	}
	return &SignedMessage{
		message: m,
		messageString: ms,
		signer: signer,
		signature: signature,
	}, nil
}
