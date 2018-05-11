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
