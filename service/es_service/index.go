package es_service

import (
	"bytes"
	"context"
	"fmt"
	"myblogx/global"
)

// 创建索引
func CreateIndex(index, mapping string) error {
	// 构建创建索引的请求体
	req := bytes.NewBufferString(mapping)

	// 调用ES的Create Index API
	res, err := global.ESClient.Indices.Create(
		index,
		global.ESClient.Indices.Create.WithBody(req),
		global.ESClient.Indices.Create.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("创建索引 %s 失败: %v", index, err)
	}
	defer res.Body.Close() // 必须关闭响应体，避免资源泄漏

	// 检查响应状态码
	if res.IsError() {
		return fmt.Errorf("创建索引 %s 失败，响应体: %s", index, res.String())
	}

	return nil
}

// 判断索引是否存在
func ExistsIndex(index string) (bool, error) {
	res, err := global.ESClient.Indices.Exists(
		[]string{index},
		global.ESClient.Indices.Exists.WithContext(context.Background()),
	)
	if err != nil {
		return false, fmt.Errorf("检查索引 %s 是否存在失败: %v", index, err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, fmt.Errorf("检查索引 %s 是否存在失败，响应状态码: %d", index, res.StatusCode)
	}
}

// 删除索引
func DeleteIndex(index string) error {
	res, err := global.ESClient.Indices.Delete(
		[]string{index},
		global.ESClient.Indices.Delete.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("删除索引 %s 失败: %v", index, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("删除索引 %s 失败，响应错误: %s", index, res.Status())
	}

	return nil
}

// 强制创建索引
func CreateIndexForce(index, mapping string) error {
	if exists, err := ExistsIndex(index); err != nil {
		return err
	} else if exists {
		if err := DeleteIndex(index); err != nil {
			return err
		}
	}
	return CreateIndex(index, mapping)
}
