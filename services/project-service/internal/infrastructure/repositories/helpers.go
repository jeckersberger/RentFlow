package repositories

import (
	"database/sql/driver"
	"encoding/json"
)

func toJSONB(data map[string]interface{}) driver.Value {
	if data == nil {
		return nil
	}
	b, _ := json.Marshal(data)
	return string(b)
}

func fromJSONB(data string) map[string]interface{} {
	var m map[string]interface{}
	if data != "" {
		json.Unmarshal([]byte(data), &m)
	}
	return m
}
