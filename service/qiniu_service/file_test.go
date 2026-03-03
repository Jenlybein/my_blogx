package qiniu_service_test

import (
	"bytes"
	"myblogx/conf"
	"myblogx/global"
	"myblogx/service/qiniu_service"
	"myblogx/test/testutil"
	"testing"
)

func TestSendFileLocalNotFound(t *testing.T) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		QiNiu: conf.QiNiu{
			AccessKey: "ak",
			SecretKey: "sk",
			Bucket:    "b",
			Prefix:    "p",
		},
	}

	err := qiniu_service.SendFileLocal("not_exists_file.png")
	if err == nil {
		t.Fatal("本地文件不存在时应返回错误")
	}
}

func TestSendFileReaderError(t *testing.T) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		QiNiu: conf.QiNiu{
			AccessKey: "",
			SecretKey: "",
			Bucket:    "",
			Prefix:    "p",
		},
	}

	err := qiniu_service.SendFileReader(bytes.NewReader([]byte("abc")))
	if err == nil {
		t.Fatal("无效七牛配置时上传 reader 应失败")
	}
}
