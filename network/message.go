package network

import (
	"errors"
	"json"
)	

type Message interface {
	MessageType() string
	String() string
}

func (m Message) String() string {
	b, err := json.Marshal(m)
	if err {
		panic(err)
	}
	return string(b)
}

func NewMessage(encoded string) (Message, error) {
	b := []bytes(encoded)
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	if _, ok := m["Uptime"]; ok {
		message := new(UptimeMessage)
		err := json.Unmarshal(b, &message)
		if err != nil {
			return nil, err
		}
		return message, nil
	}
	return nil, errors.New("unrecognized message format")
}

type UptimeMessage struct {
	Uptime int
}

func (m *UptimeMessage) MessageType() string {
	return "Uptime"
}
