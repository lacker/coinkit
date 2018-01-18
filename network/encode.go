package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"coinkit/consensus"
	"coinkit/util"
)

func EncodeMessage(m util.Message) string {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func DecodeMessage(encoded string) (util.Message, error) {
	b := []byte(encoded)
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	if _, ok := m["Acc"]; ok {
		message := new(consensus.NominationMessage)
		err := json.Unmarshal(b, &message)
		if err != nil {
			return nil, err
		}
		return message, nil
	}
	if messageType, ok := m["T"]; ok {
		switch mt := messageType.(type) {
		case float64:
			switch consensus.Phase(mt) {
			case consensus.Prepare:
				message := new(consensus.PrepareMessage)
				err := json.Unmarshal(b, &message)
				if err != nil {
					return nil, err
				}
				return message, nil
			case consensus.Confirm:
				message := new(consensus.ConfirmMessage)
				err := json.Unmarshal(b, &message)
				if err != nil {
					return nil, err
				}
				return message, nil
			case consensus.Externalize:
				message := new(consensus.ExternalizeMessage)
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

// Useful for simulating a network transit
func EncodeThenDecode(message util.Message) util.Message {
	encoded := EncodeMessage(message)
	m, err := DecodeMessage(encoded)
	if err != nil {
		log.Fatal("encode-then-decode error:", err)
	}
	return m
}
