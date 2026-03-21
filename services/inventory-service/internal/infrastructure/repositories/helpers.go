package repositories

import (
	"encoding/json"

	"github.com/lib/pq"
)

// toJSONB converts a map to JSONB byte array
func toJSONB(data map[string]string) []byte {
	if data == nil {
		data = make(map[string]string)
	}
	jsonBytes, _ := json.Marshal(data)
	return jsonBytes
}

// fromJSONB converts JSONB byte array to map
func fromJSONB(data []byte) map[string]string {
	result := make(map[string]string)
	if data == nil || len(data) == 0 {
		return result
	}
	json.Unmarshal(data, &result)
	return result
}

// pq is imported from lib/pq for array types
var _ = pq.StringArray{}
