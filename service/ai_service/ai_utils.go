package ai_service

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

var (
	jsonFenceRegexp = regexp.MustCompile("(?is)```(?:json)?\\s*(\\{.*?\\}|\\[.*?\\])\\s*```")
	jsonBodyRegexp  = regexp.MustCompile("(?is)(\\{.*\\}|\\[.*\\])")
)

// MustJSONString 把任意值序列化为 JSON 字符串；失败时返回空数组字面量。
func MustJSONString(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// ExtractJSONBlock 从 AI 返回文本中提取首个 JSON 块。
// 这里优先提取 ```json ... ``` 代码块，其次再回退到普通文本里的 JSON 片段。
func ExtractJSONBlock(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("AI 返回内容为空")
	}

	if match := jsonFenceRegexp.FindStringSubmatch(text); len(match) >= 2 {
		return strings.TrimSpace(match[1]), nil
	}

	if match := jsonBodyRegexp.FindStringSubmatch(text); len(match) >= 2 {
		return strings.TrimSpace(match[1]), nil
	}

	return "", errors.New("未找到有效 JSON 内容")
}

// UnmarshalJSONBlock 先提取 JSON 块，再反序列化到目标对象。
func UnmarshalJSONBlock(text string, target any) error {
	raw, err := ExtractJSONBlock(text)
	if err != nil {
		return err
	}
	if err = json.Unmarshal([]byte(raw), target); err != nil {
		return err
	}
	return nil
}
