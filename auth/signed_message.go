package auth

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
