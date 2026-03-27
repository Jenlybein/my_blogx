package image_service

import (
	"bytes"
	"encoding/base64"
	"mime/multipart"
	"myblogx/conf"
	"myblogx/global"
	"myblogx/test/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreateUploadToken(t *testing.T) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		QiNiu: conf.QiNiu{
			AccessKey: "ak",
			SecretKey: "sk",
			Bucket:    "bucket",
		},
	}

	ret, err := CreateUploadToken(UploadPolicy{
		Bucket:      "bucket",
		ObjectKey:   "blogx/images/test.png",
		CallbackURL: "https://api.example.com/api/images/qiniu/callback",
		ExpireAt:    time.Now().Add(time.Hour),
		MaxSize:     5 * 1024 * 1024,
		EndUser:     "1",
	})
	if err != nil {
		t.Fatalf("CreateUploadToken 失败: %v", err)
	}
	if ret == nil || ret.Token == "" {
		t.Fatal("上传 token 不应为空")
	}
	if !strings.Contains(ret.Token, ":") {
		t.Fatalf("token 格式异常: %s", ret.Token)
	}
}

func TestCreateUploadTokenWithoutCallback(t *testing.T) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		QiNiu: conf.QiNiu{
			AccessKey: "ak",
			SecretKey: "sk",
			Bucket:    "bucket",
		},
	}

	ret, err := CreateUploadToken(UploadPolicy{
		Bucket:    "bucket",
		ObjectKey: "blogx/images/test-no-callback.png",
		ExpireAt:  time.Now().Add(time.Hour),
		MaxSize:   5 * 1024 * 1024,
		EndUser:   "1",
	})
	if err != nil {
		t.Fatalf("未配置回调地址时也应允许签发上传 token: %v", err)
	}
	if ret == nil || ret.Token == "" {
		t.Fatal("上传 token 不应为空")
	}
}

func TestCreateUploadTokenInvalidPolicy(t *testing.T) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		QiNiu: conf.QiNiu{
			AccessKey: "ak",
			SecretKey: "sk",
			Bucket:    "bucket",
		},
	}

	ret, err := CreateUploadToken(UploadPolicy{
		Bucket:      "",
		ObjectKey:   "blogx/images/test.png",
		CallbackURL: "https://api.example.com/api/images/qiniu/callback",
		ExpireAt:    time.Now().Add(time.Hour),
		MaxSize:     5 * 1024 * 1024,
	})
	if err == nil || ret != nil {
		t.Fatalf("非法策略应失败, ret=%+v err=%v", ret, err)
	}
}

func TestImageSuffixAndVerifyFormat(t *testing.T) {
	testutil.InitGlobals()

	if s := GetImageSuffix("a.JPG"); s != "jpg" {
		t.Fatalf("GetImageSuffix 错误: %s", s)
	}
	if s := GetImageSuffix("noext"); s != "" {
		t.Fatalf("无后缀应返回空: %s", s)
	}

	pngData, _ := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO7+f2UAAAAASUVORK5CYII=",
	)
	h := makeMultipartImageHeader(t, "x.png", pngData)

	err := VerifyImageFormat([]string{"png", "jpg"}, h)
	if err != nil {
		t.Fatalf("合法图片校验失败: %v", err)
	}

	bad := makeMultipartImageHeader(t, "x.jpg", pngData)
	if err = VerifyImageFormat([]string{"jpg"}, bad); err == nil {
		t.Fatal("后缀与内容不匹配应报错")
	}
}

func makeMultipartImageHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile 失败: %v", err)
	}
	if _, err = part.Write(content); err != nil {
		t.Fatalf("写入 multipart 内容失败: %v", err)
	}
	if err = w.Close(); err != nil {
		t.Fatalf("关闭 multipart writer 失败: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if err = req.ParseMultipartForm(8 << 20); err != nil {
		t.Fatalf("ParseMultipartForm 失败: %v", err)
	}

	return req.MultipartForm.File["file"][0]
}
