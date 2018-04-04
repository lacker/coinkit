package data

import (
	"encoding/json"
)

type Document struct {
	// For consistency, all fields on a document are stored within the
	// Data column.
	// Naming convention is namedLikeThis.
	// Some fields are required on every object:
	// id: a unique integer
	// TODO: collection, owner, createdAt, updatedAt
	Data map[string]interface{}

	// Every document has a unique id. It is stored twice in the
	// database to enforce uniqueness.
	Id uint64
}

func (d *Document) String() string {
	bytes, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(append(bytes, '\n'))
}

func NewDocument(id uint64) *Document {
	return &Document{
		Data: map[string]interface{}{
			"id": id,
		},
		Id: id,
	}
}
