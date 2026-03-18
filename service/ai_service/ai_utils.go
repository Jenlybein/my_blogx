package ai_service

import "encoding/json"

func mustJSONString(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}
