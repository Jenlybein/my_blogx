package core

import (
	"context"

	"myblogx/global"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
)

// EsConnect 初始化并连接 Elasticsearch（修复上下文传参错误）
func EsConnect() *elasticsearch.Client {
	esConf := global.Config.ES
	if esConf.Addresses == nil {
		logrus.Info("ES 地址配置为空，不启用")
		return nil
	}

	// 配置客户端参数
	cfg := elasticsearch.Config{
		Addresses: esConf.Addresses,
		Username:  esConf.Username,
		Password:  esConf.Password,
	}

	// 创建客户端实例
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logrus.Fatalf("创建 ES 客户端失败: %v", err)
		return nil
	}

	// 3. 验证连接（修正上下文传参方式）
	ctx := context.Background()
	resp, err := es.Info(es.Info.WithContext(ctx))
	if err != nil {
		logrus.Fatalf("验证 ES 连接失败: %v", err)
		return nil
	}
	defer resp.Body.Close() // 必须关闭响应体

	// 检查响应状态
	if resp.IsError() {
		logrus.Fatalf("ES 连接异常，状态码: %s", resp.Status())
		return nil
	}

	// 4. 赋值全局客户端
	logrus.Infof("ES 客户端连接成功：%s", esConf.Addresses)
	return es
}
