package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONObject is a modifiable set of key-value mappings.
// It is designed to be more convenient than representing JSON as either a
// map[string]interface{} or the raw bytes.
// After calling any exposed method, bytes and content should be equivalent.
type JSONObject struct {
	bytes   []byte
	content map[string]interface{}
}

// Sets bytes based on content
func (ob *JSONObject) encode() {
	bytes, err := json.Marshal(ob.content)
	if err != nil {
		panic(err)
	}
	ob.bytes = bytes
}

// Sets content based on bytes.
// Returns an error if the bytes are bad json
func (ob *JSONObject) decode() error {
	return json.Unmarshal(ob.bytes, &ob.content)
}

func NewJSONObject(content map[string]interface{}) *JSONObject {
	answer := &JSONObject{
		content: content,
	}
	answer.encode()
	return answer
}

func NewEmptyJSONObject() *JSONObject {
	content := make(map[string]interface{})
	return NewJSONObject(content)
}

func (ob *JSONObject) Value() (driver.Value, error) {
	return driver.Value(ob.bytes), nil
}

func (ob *JSONObject) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return errors.New("expected []byte")
	}
	ob.bytes = bytes
	return json.Unmarshal(ob.bytes, &ob.content)
}
