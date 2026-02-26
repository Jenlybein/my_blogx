package file_test

import (
	"bytes"
	"encoding/base64"
	"mime/multipart"
	"myblogx/test/testutil"
	"myblogx/utils/file"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImageSuffixAndVerifyFormat(t *testing.T) {
	testutil.InitGlobals()

	if s := file.GetImageSuffix("a.JPG"); s != "jpg" {
		t.Fatalf("GetImageSuffix 错误: %s", s)
	}
	if s := file.GetImageSuffix("noext"); s != "" {
		t.Fatalf("无后缀应返回空: %s", s)
	}

	pngData, _ := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO7+f2UAAAAASUVORK5CYII=",
	)
	h := makeMultipartImageHeader(t, "x.png", pngData)

	err := file.VerifyImageFormat([]string{"png", "jpg"}, h)
	if err != nil {
		t.Fatalf("合法图片校验失败: %v", err)
	}

	bad := makeMultipartImageHeader(t, "x.jpg", pngData)
	if err = file.VerifyImageFormat([]string{"jpg"}, bad); err == nil {
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
