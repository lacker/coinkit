package data

import (
	"fmt"

	"coinkit/util"
)

// A DataMessage is used to exchange data among peers.
// An empty string means that the sender has this value, it just isn't included.
type DataMessage struct {
	Data map[string]string
}

func (m *DataMessage) Slot() int {
	return 0
}

func (m *DataMessage) MessageType() string {
	return "D"
}

func (m *DataMessage) String() string {
	plural := ""
	if len(m.Data) == 1 {
		plural = "s"
	}
	return fmt.Sprintf("data(%d key%s)", len(m.Data), plural)
}

func init() {
	util.RegisterMessageType(&DataMessage{})
}
