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

// Returns a descriptive error if this is not canonical json
func CheckCanonicalJSON(bs []byte) error {
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

// JSON-encodes something in a canonical way.
// The optional choices that make this canonical:
//   * Field order is alphabetized
//   * & < > characters are not escaped
// TODO: this encodes twice. find a more efficient way to do this
func CanonicalJSONEncode(x interface{}) []byte {
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
