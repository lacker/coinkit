package data

import ()

// A DocumentQuery expresses a way to select a subset of documents.
type DocumentQuery struct {
	// Each field in data expresses an exact match with the data of the object to
	// be retrieved.
	Data *JSONObject

	// The maximum number of objects to be returned.
	// It's up to individual servers what the maximum supported limit is.
	Limit int
}
