package util

import (
	"encoding/json"
)

func PrettyJSON(x interface{}) []byte {
	bytes, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(bytes, '\n')
}

func ToJSON(x interface{}) []byte {
	bytes, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return bytes
}

// JSON-encodes something, and also alphabetizes the fields.
// TODO: this encodes twice. find a more efficient way to do this
func AlphabeticalJSONEncode(x interface{}) []byte {
	encoded, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	var decoded interface{}
	json.Unmarshal(encoded, &decoded)
	reencoded, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return reencoded
}
