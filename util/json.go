package util

import (
	"bytes"
	"encoding/json"
)

func PrettyJSON(x interface{}) []byte {
	bs, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(bs, '\n')
}

func ToJSON(x interface{}) []byte {
	bs, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return bs
}

func IsAlphabeticalJSON(bs []byte) bool {
	var decoded interface{}
	json.Unmarshal(bs, &decoded)
	reencoded, err := json.Marshal(decoded)
	if err != nil {
		return false
	}
	return bytes.Compare(bs, reencoded) == 0
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
