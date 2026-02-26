package es_service

import (
	"io"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func TestBuildBulkBody(t *testing.T) {
	items := []*BulkRequest{
		{
			Action: ActionIndex,
			Index:  "idx",
			ID:     "1",
			Data: map[string]interface{}{
				"title": "hello",
			},
		},
		{
			Action: ActionUpdate,
			Index:  "idx",
			ID:     "2",
			Data: map[string]interface{}{
				"title": "world",
			},
		},
		{
			Action: ActionDelete,
			Index:  "idx",
			ID:     "3",
		},
	}

	body, err := buildBulkBody(items)
	if err != nil {
		t.Fatalf("buildBulkBody 失败: %v", err)
	}
	s := string(body)
	if !strings.Contains(s, "\"index\"") || !strings.Contains(s, "\"update\"") || !strings.Contains(s, "\"delete\"") {
		t.Fatalf("bulk body 缺少 action: %s", s)
	}
	if !strings.Contains(s, "\"doc\"") {
		t.Fatalf("update 文档结构缺失: %s", s)
	}
}

func TestHandleError(t *testing.T) {
	res := &esapi.Response{
		StatusCode: 400,
		Body: io.NopCloser(strings.NewReader(
			`{"error":{"reason":"bad request","caused_by":{"reason":"x"}}}`,
		)),
	}
	err := handleError(res)
	if err == nil {
		t.Fatal("handleError 应返回错误")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("错误信息异常: %v", err)
	}
}

func TestCloseResponse(t *testing.T) {
	res := &esapi.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}
	closeResponse(res)
}
