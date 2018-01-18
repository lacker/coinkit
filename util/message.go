package util

type Message interface {
	MessageType() string

	// Slot() returns 0 if the message doesn't relate to a particular slot
	Slot() int
}
