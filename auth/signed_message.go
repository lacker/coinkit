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
		signature: kp.Sign(message).
	}
}
