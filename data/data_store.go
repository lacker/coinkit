package data

import (
	"log"
)

// The DataStore stores data in memory. It's very simple now but could become more
// complicated. It needs to be more complicated to prevent abuse.
type DataStore struct {
	store map[string]string
}

func NewDataStore() *DataStore {
	return &DataStore{
		store: make(map[string]string),
	}
}

// Adds any data in the message to our store, and returns any data the sender was
// missing.
func (d *DataStore) Handle(m *DataMessage) *DataMessage {
	if m == nil || m.Data == nil {
		return nil
	}

	// Adds data to our store
	for key, value := range m.Data {
		if value == "" {
			continue
		}
		if _, ok := d.store[key]; !ok {
			log.Printf("received %s -> %d bytes", key, len(value))
			d.store[key] = value
		}
	}

	// Gives them the data they are missing
	answer := &DataMessage{
		Data: make(map[string]string),
	}
	count := 0
	for key, value := range d.store {
		if _, ok := m.Data[key]; !ok {
			answer.Data[key] = value
			count += 1
		} else {
			answer.Data[key] = ""
		}
	}
	if count == 0 {
		return nil
	}
	return answer
}

func (d *DataStore) DataMessage() *DataMessage {
	message := &DataMessage{
		Data: make(map[string]string),
	}
	for key, _ := range d.store {
		message.Data[key] = ""
	}
	return message
}
