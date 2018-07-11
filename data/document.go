package data

import (
	"github.com/lacker/coinkit/util"
)

// TODO: reconsider whether ids should be stored twice.
type Document struct {
	// For consistency, all fields on a document are stored within the
	// Data column.
	// Naming convention is namedLikeThis.
	// Some fields are required on every object in the database, and
	// automatically added on object creation:
	// id: a unique integer
	// owner: the public key of the account that owns this document
	// TODO: consider collection, createdAt, updatedAt
	Data *JSONObject

	// Every document has a unique id, starting at 1. It is stored twice in the
	// database to enforce uniqueness.
	Id uint64
}

func (d *Document) String() string {
	return string(util.PrettyJSON(d))
}

func NewDocument(id uint64, data map[string]interface{}) *Document {
	fullData := map[string]interface{}{"id": id}
	for key, value := range data {
		fullData[key] = value
	}

	return &Document{
		Data: NewJSONObject(fullData),
		Id:   id,
	}
}
