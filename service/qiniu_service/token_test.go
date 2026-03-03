package qiniu_service_test

import (
	"myblogx/conf"
	"myblogx/global"
	"myblogx/service/qiniu_service"
	"myblogx/test/testutil"
	"strings"
	"testing"
)

func TestGenUpToken(t *testing.T) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		QiNiu: conf.QiNiu{
			AccessKey: "ak",
			SecretKey: "sk",
			Bucket:    "bucket",
			Expiry:    3600,
		},
	}

	token, err := qiniu_service.GenUpToken("test-key")
	if err != nil {
		t.Fatalf("GenUpToken 失败: %v", err)
	}
	if token == "" {
		t.Fatal("token 不应为空")
	}
	if !strings.Contains(token, ":") {
		t.Fatalf("token 格式异常: %s", token)
	}
}

func TestGenUpTokenInvalidConfig(t *testing.T) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		QiNiu: conf.QiNiu{
			AccessKey: "ak",
			SecretKey: "sk",
			Bucket:    "",
			Expiry:    3600,
		},
	}

	token, err := qiniu_service.GenUpToken("test-key")
	if err == nil {
		t.Fatalf("bucket 为空时应失败, token=%s", token)
	}
}
