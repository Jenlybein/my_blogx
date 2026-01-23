package es_service

import (
	"bytes"
	"context"
	"fmt"
	"myblogx/global"
)

// 创建 pipeline
func CreatePipeline(pipeline, definition string) error {
	// 构建创建 pipeline 的请求体
	req := bytes.NewBufferString(definition)

	// 调用 ES 的 Create Pipeline API
	res, err := global.ESClient.Ingest.PutPipeline(
		pipeline,
		req,
		global.ESClient.Ingest.PutPipeline.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("创建 pipeline %s 失败: %v", pipeline, err)
	}
	defer res.Body.Close() // 必须关闭响应体，避免资源泄漏

	// 检查响应状态码
	if res.IsError() {
		return fmt.Errorf("创建 pipeline %s 失败，响应体: %s", pipeline, res.String())
	}

	return nil
}

// 判断 pipeline 是否存在
func ExistsPipeline(pipeline string) (bool, error) {
	// 必须指定 PipelineID 才能查询特定的 pipeline
	res, err := global.ESClient.Ingest.GetPipeline(
		global.ESClient.Ingest.GetPipeline.WithPipelineID(pipeline), // 关键点：指定 ID
		global.ESClient.Ingest.GetPipeline.WithContext(context.Background()),
	)

	if err != nil {
		return false, fmt.Errorf("检查 pipeline %s 发生错误: %v", pipeline, err)
	}
	defer res.Body.Close()

	// 如果状态码是 404，说明该特定的 pipeline 不存在
	if res.StatusCode == 404 {
		return false, nil
	}

	// 如果有其他错误（如 500），记录日志并返回 false
	if res.IsError() {
		return false, fmt.Errorf("检查 pipeline %s 失败，响应状态码: %d", pipeline, res.StatusCode)
	}

	// 状态码为 200，说明存在
	return true, nil
}

// 删除 pipeline
func DeletePipeline(pipeline string) error {
	res, err := global.ESClient.Ingest.DeletePipeline(
		pipeline,
		global.ESClient.Ingest.DeletePipeline.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("删除 pipeline %s 失败: %v", pipeline, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("管道 %s 删除失败，响应错误: %s", pipeline, res.Status())
	}

	return nil
}

// 强制创建 pipeline
func CreatePipelineForce(pipeline, definition string) error {
	if exists, err := ExistsPipeline(pipeline); err != nil {
		return err
	} else if exists {
		if err := DeletePipeline(pipeline); err != nil {
			return err
		}
	}
	return CreatePipeline(pipeline, definition)
}
