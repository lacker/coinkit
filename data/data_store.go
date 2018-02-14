package data

import (
	"log"
	"sync"
)

// The DataStore stores data in memory. It's very simple now but could become more
// complicated. It needs to be more complicated to prevent abuse.
// The DataStore is threadsafe.
type DataStore struct {
	store map[string]string
	mutex *sync.Mutex
}

func NewDataStore() *DataStore {
	return &DataStore{
		store: make(map[string]string),
		mutex: &sync.Mutex{},
	}
}

// Adds any data in the message to our store, and returns any data the sender was
// missing.
func (d *DataStore) Handle(m *DataMessage) *DataMessage {
	d.mutex.Lock()
	defer d.mutex.Unlock()

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
		Open: 0,
	}
	bytesToSend := 0
	for key, value := range d.store {
		if _, ok := m.Data[key]; ok {
			continue
		}
		bytesToSend += len(key) + len(value)
		if bytesToSend > m.Open {
			break
		}
		answer.Data[key] = value
	}
	if len(answer.Data) == 0 {
		return nil
	}
	return answer
}

func (d *DataStore) DataMessage() *DataMessage {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Request at most 5-megabyte files.
	// TODO: better enforcement of space-used
	message := &DataMessage{
		Data: make(map[string]string),
		Open: 5000000,
	}
	for key, _ := range d.store {
		message.Data[key] = ""
	}
	return message
}

func (d *DataStore) Get(key string) (string, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	value, ok := d.store[key]
	return value, ok
}
