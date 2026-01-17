package es_service

import (
	"bytes"
	"context"
	"myblogx/global"
)

// 创建索引
func CreateIndex(index, mapping string) {
	// 构建创建索引的请求体
	req := bytes.NewBufferString(mapping)

	// 调用ES的Create Index API
	res, err := global.ESClient.Indices.Create(
		index,
		global.ESClient.Indices.Create.WithBody(req),
		global.ESClient.Indices.Create.WithContext(context.Background()),
	)
	if err != nil {
		global.Logger.Errorf("创建索引 %s 失败: %v", index, err)
		return
	}
	defer res.Body.Close() // 必须关闭响应体，避免资源泄漏

	// 检查响应状态码
	if res.IsError() {
		global.Logger.Errorf("创建索引 %s 失败，响应错误: %s", index, res.Status())
		return
	}

	global.Logger.Infof("索引 %s 创建成功", index)
}

// 判断索引是否存在
func ExistsIndex(index string) bool {
	res, err := global.ESClient.Indices.Exists(
		[]string{index},
		global.ESClient.Indices.Exists.WithContext(context.Background()),
	)
	if err != nil {
		global.Logger.Errorf("检查索引 %s 是否存在失败: %v", index, err)
		return false
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		return true
	case 404:
		return false
	default:
		global.Logger.Errorf("检查索引 %s 是否存在失败，响应状态码: %d", index, res.StatusCode)
		return false
	}
}

// 删除索引
func DeleteIndex(index string) {
	res, err := global.ESClient.Indices.Delete(
		[]string{index},
		global.ESClient.Indices.Delete.WithContext(context.Background()),
	)
	if err != nil {
		global.Logger.Errorf("删除索引 %s 失败: %v", index, err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		global.Logger.Errorf("删除索引 %s 失败，响应错误: %s", index, res.Status())
		return
	}

	global.Logger.Infof("索引 %s 删除成功", index)
}

func CreateIndexForce(index, mapping string) {
	if ExistsIndex(index) {
		DeleteIndex(index)
	}
	CreateIndex(index, mapping)
}
