package ai_service

import "encoding/json"

// MustJSONString 把任意值序列化为 JSON 字符串；失败时返回空数组字面量。
func MustJSONString(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}
