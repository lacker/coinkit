package network

import (
	"encoding/json"
	"errors"
)	

type Message interface {
	MessageType() string
}

func EncodeMessage(m Message) string {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func DecodeMessage(encoded string) (Message, error) {
	b := []byte(encoded)
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
	if _, ok := m["X"]; ok {
		message := new(NominationMessage)
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
