package repositories

import (
	"encoding/json"
	"github.com/lib/pq"
)

func toStringArray(s []string) pq.StringArray {
	if len(s) == 0 {
		return pq.StringArray{}
	}
	return pq.StringArray(s)
}

func fromStringArray(arr pq.StringArray) []string {
	if arr == nil {
		return []string{}
	}
	return []string(arr)
}

func toJSONB(data map[string]string) []byte {
	if len(data) == 0 {
		return []byte("{}")
	}
	b, _ := json.Marshal(data)
	return b
}
