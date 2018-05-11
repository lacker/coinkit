package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/lacker/coinkit/util"
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

func (ob *JSONObject) setBytes(bytes []byte) error {
	ob.bytes = bytes
	return json.Unmarshal(ob.bytes, &ob.content)
}

func NewJSONObject(content map[string]interface{}) *JSONObject {
	answer := &JSONObject{
		content: content,
	}
	answer.encode()
	return answer
}

func ReadJSONObject(bytes []byte) (*JSONObject, error) {
	ob := &JSONObject{}
	err := ob.setBytes(bytes)
	if err == nil {
		return ob, nil
	}
	return nil, err
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
	return ob.setBytes(bytes)
}

func (ob *JSONObject) Set(key string, value interface{}) {
	ob.content[key] = value
	ob.encode()
}

// Returns (nil, false) if the key does not exist
func (ob *JSONObject) Get(key string) (interface{}, bool) {
	value, ok := ob.content[key]
	return value, ok
}

// Returns (0, false) if the key does not exist, or is not an int
func (ob *JSONObject) GetInt(key string) (int, bool) {
	value, ok := ob.Get(key)
	if ok {
		intValue, ok := value.(int)
		if ok {
			return intValue, true
		}
	}
	return 0, false
}

func (ob *JSONObject) String() string {
	return string(util.PrettyJSON(ob))
}
