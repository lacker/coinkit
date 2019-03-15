package util

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// Returns a descriptive error if this is not alphabetical json
func CheckAlphabeticalJSON(bs []byte) error {
	var decoded interface{}
	json.Unmarshal(bs, &decoded)
	reencoded, err := json.Marshal(decoded)
	if err != nil {
		return fmt.Errorf("could not reencode json: %s", err)
	}
	if bytes.Compare(bs, reencoded) != 0 {
		return fmt.Errorf("unalphabetical json:\n%s\nthe alphabetized version is:\n%s",
			bs, reencoded)
	}
	return nil
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
