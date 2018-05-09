package data

import (
	"encoding/json"
)

// JSONObject is a modifiable set of key-value mappings.
// It is designed to be more convenient than representing JSON as either a
// map[string]interface{} or the raw bytes.
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

func NewJSONObject() *JSONObject {
	answer := &JSONObject{
		content: make(map[string]interface{}),
	}
	answer.encode()
	return answer
}
