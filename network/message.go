package network

import (
	"encoding/json"
	"errors"
	"fmt"
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
	if _, ok := m["Acc"]; ok {
		message := new(NominationMessage)
		err := json.Unmarshal(b, &message)
		if err != nil {
			return nil, err
		}
		return message, nil
	}
	if messageType, ok := m["T"]; ok {
		switch mt := messageType.(type) {
		case float64:
			switch Phase(mt) {
			case Prepare:
				message := new(PrepareMessage)
				err := json.Unmarshal(b, &message)
				if err != nil {
					return nil, err
				}
				return message, nil
			case Confirm:
				message := new(ConfirmMessage)
				err := json.Unmarshal(b, &message)
				if err != nil {
					return nil, err
				}
				return message, nil
			case Externalize:
				message := new(ExternalizeMessage)
				err := json.Unmarshal(b, &message)
				if err != nil {
					return nil, err
				}
				return message, nil
			default:
				return nil, fmt.Errorf("bad ballot phase: %v", messageType)
			}
		default:
			return nil, fmt.Errorf("bad T: %#v", messageType)
		}
	}
	return nil, errors.New("unrecognized message format")
}
