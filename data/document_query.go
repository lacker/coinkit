package data

import ()

// A DocumentQuery expresses a way to select a subset of documents.
type DocumentQuery struct {
	// Each field in data expresses an exact match with the data of the object to
	// be retrieved.
	Data *JSONObject

	// The maximum number of objects to be returned.
	Limit int
}
