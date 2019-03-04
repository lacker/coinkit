package util

import (
	"bytes"
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

func IsAlphabeticalJSON(bytes []byte) bool {
	var decoded interface{}
	json.Unmarshal(bytes, &decoded)
	reencoded, err := json.Marshal(decoded)
	if err != nil {
		return false
	}
	return bytes.Compare(bytes, reencoded) == 0
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
	reencoded, err := json.Marshal(decoded)
	if err != nil {
		panic(err)
	}
	return reencoded
}
