package es_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"myblogx/global"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// 通用响应结构体
type ESResponse struct {
	Success bool   // 操作是否成功
	Msg     string // 提示信息
	Data    any    // 响应数据（文档ID/列表/版本号等）
}

// --- 通用工具函数 ---

// 安全关闭响应体，避免资源泄漏
func closeResponse(res *esapi.Response) {
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}

// 错误处理，提取 ES 错误信息
func handleError(res *esapi.Response) error {
	var errResp map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&errResp); decodeErr != nil {
		return fmt.Errorf("解析错误响应失败: %v，状态: %s", decodeErr, res.Status())
	}

	errorReason := ""

	// 尝试从 ES 标准错误对象中提取 reason
	if errorObj, ok := errResp["error"].(map[string]any); ok {
		if reason, ok := errorObj["reason"].(string); ok {
			errorReason = reason
		}
		// 进一步检查是否有更深层的错误原因（如 causado_by）
		if causedBy, ok := errorObj["caused_by"].(map[string]any); ok {
			if cbReason, ok := causedBy["reason"].(string); ok {
				errorReason = fmt.Sprintf("%s (原因: %s)", errorReason, cbReason)
			}
		}
	}

	// 如果没能提取到有效 Reason，根据状态码提供常见解释
	if errorReason == "" {
		switch res.StatusCode {
		case 400:
			errorReason = "请求参数错误"
		case 401:
			errorReason = "未授权或凭证失效"
		case 403:
			errorReason = "权限不足，拒绝访问"
		case 404:
			errorReason = "资源未找到"
		case 409:
			errorReason = "版本冲突或文档已存在"
		default:
			errorReason = "未知错误"
		}
	}

	return fmt.Errorf("%s (状态码: %d)", errorReason, res.StatusCode)
}

// 执行ES请求，处理错误响应
func doRequest(req esapi.Request) (res *esapi.Response, err error) {
	res, err = req.Do(context.Background(), global.ESClient)
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		defer closeResponse(res)
		return nil, handleError(res)
	}
	return res, nil
}

// 泛型解析响应体
func decodeResponse(body io.ReadCloser) (map[string]any, error) {
	var target map[string]any
	defer body.Close()
	err := json.NewDecoder(body).Decode(&target)
	return target, err
}

// --- 业务操作封装 ---

// 创建通用文档（泛型）
func CreateDocument(index string, data any) ESResponse {
	body, _ := json.Marshal(data)
	req := esapi.IndexRequest{
		Index: index,
		Body:  bytes.NewReader(body),
	}

	res, err := doRequest(req)
	if err != nil {
		return ESResponse{Success: false, Msg: err.Error()}
	}
	defer closeResponse(res)

	return ESResponse{Success: true, Msg: "创建成功", Data: res.String()}
}

// Search 分页查询封装
func Search[T any](index string, page, limit int, query map[string]any) ESResponse {
	from := (page - 1) * limit
	searchBody := map[string]any{"from": from, "size": limit, "query": query}
	body, _ := json.Marshal(searchBody)

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(body),
	}

	res, err := doRequest(req)
	if err != nil {
		return ESResponse{Success: false, Msg: err.Error()}
	}

	// 这里的 map[string]any 可以根据 ES 返回结构定义更精确的 struct
	result, _ := decodeResponse(res.Body)
	hitsObj := result["hits"].(map[string]any)

	return ESResponse{
		Success: true,
		Msg:     "查询成功",
		Data: map[string]any{
			"total": hitsObj["total"].(map[string]any)["value"],
			"hits":  hitsObj["hits"], // 保持原始 hits 供上层解析，或在此处进一步泛型转换
		},
	}
}

// 更新通用文档
func UpdateDocument(index, docID string, updateData map[string]any) ESResponse {
	body, _ := json.Marshal(map[string]any{"doc": updateData})
	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: docID,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := doRequest(req)
	if err != nil {
		return ESResponse{Success: false, Msg: err.Error()}
	}

	result, _ := decodeResponse(res.Body)
	return ESResponse{Success: true, Msg: "更新成功", Data: result}
}

// DeleteDocument 删除通用文档
func DeleteDocument(index, docID string) ESResponse {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: docID,
		Refresh:    "true",
	}

	res, err := doRequest(req)
	if err != nil {
		return ESResponse{Success: false, Msg: err.Error()}
	}

	result, _ := decodeResponse(res.Body)
	return ESResponse{Success: true, Msg: "删除成功", Data: result}
}
