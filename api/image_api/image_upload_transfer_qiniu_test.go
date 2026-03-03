package image_api_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"myblogx/api/image_api"
	"myblogx/conf"
	"myblogx/global"
	"myblogx/models"
	"myblogx/test/testutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

var onePixelPNG = mustDecodeBase64("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO6p7N8AAAAASUVORK5CYII=")

func mustDecodeBase64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func readCode2(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v body=%s", err, w.Body.String())
	}
	return int(body["code"].(float64))
}

func newImageCtxWithRequest(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

func newMultipartRequest(t *testing.T, field, name string, content []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(field, name)
	if err != nil {
		t.Fatalf("创建 multipart file 失败: %v", err)
	}
	if _, err = part.Write(content); err != nil {
		t.Fatalf("写入 multipart 内容失败: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("关闭 multipart writer 失败: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func setupImageUploadEnv(t *testing.T, uploadDir string) {
	t.Helper()
	testutil.SetupSQLite(t, &models.ImageModel{})
	testutil.InitGlobals()
	global.Config = &conf.Config{
		Upload: conf.Upload{
			Whitelist: []string{"png", "jpg", "jpeg"},
			UploadDir: uploadDir,
		},
	}
	if err := os.MkdirAll(filepath.Join("uploads", uploadDir), 0755); err != nil {
		t.Fatalf("创建上传目录失败: %v", err)
	}
}

func TestImageUploadView(t *testing.T) {
	uploadDir := "unit_upload_api"
	setupImageUploadEnv(t, uploadDir)
	api := &image_api.ImageApi{}

	t.Run("超过 2MB", func(t *testing.T) {
		tooLarge := bytes.Repeat([]byte("a"), 2*1024*1024+1)
		req := newMultipartRequest(t, "file", "big.png", tooLarge)
		c, w := newImageCtxWithRequest(req)
		api.ImageUploadView(c)
		if code := readCode2(t, w); code == 0 {
			t.Fatalf("超大文件应失败, body=%s", w.Body.String())
		}
	})

	t.Run("上传成功并去重", func(t *testing.T) {
		req := newMultipartRequest(t, "file", "a.png", onePixelPNG)
		c, w := newImageCtxWithRequest(req)
		api.ImageUploadView(c)
		if code := readCode2(t, w); code != 0 {
			t.Fatalf("首次上传应成功, body=%s", w.Body.String())
		}

		var cnt int64
		_ = global.DB.Model(&models.ImageModel{}).Count(&cnt).Error
		if cnt != 1 {
			t.Fatalf("上传成功后应有 1 条记录, cnt=%d", cnt)
		}

		req2 := newMultipartRequest(t, "file", "b.png", onePixelPNG)
		c2, w2 := newImageCtxWithRequest(req2)
		api.ImageUploadView(c2)
		if code := readCode2(t, w2); code == 0 {
			t.Fatalf("重复文件应失败, body=%s", w2.Body.String())
		}
	})
}

func TestTransferSaveView(t *testing.T) {
	uploadDir := "unit_transfer_api"
	setupImageUploadEnv(t, uploadDir)
	api := &image_api.ImageApi{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(onePixelPNG)
	}))
	defer srv.Close()

	c, w := newImageCtxWithRequest(httptest.NewRequest(http.MethodPost, "/transfer", nil))
	c.Set("requestJson", image_api.TransferSaveRequest{ImageURL: srv.URL})
	api.TransferSaveView(c)
	if code := readCode2(t, w); code != 0 {
		t.Fatalf("TransferSaveView 应成功, body=%s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "/uploads/"+uploadDir+"/") {
		t.Fatalf("返回路径异常: %s", w.Body.String())
	}
}

func TestGenUpTokenView(t *testing.T) {
	testutil.InitGlobals()
	api := &image_api.ImageApi{}

	global.Config = &conf.Config{QiNiu: conf.QiNiu{Enable: false}}
	c, w := newImageCtxWithRequest(httptest.NewRequest(http.MethodGet, "/token", nil))
	api.GenUpToken(c)
	if code := readCode2(t, w); code == 0 {
		t.Fatalf("七牛未启用应失败, body=%s", w.Body.String())
	}

	global.Config = &conf.Config{
		QiNiu: conf.QiNiu{
			Enable:    true,
			AccessKey: "ak",
			SecretKey: "sk",
			Bucket:    "bucket",
			Uri:       "https://cdn.example.com",
			Region:    "z0",
			Prefix:    "blogx",
			Size:      2048,
			Expiry:    3600,
		},
	}
	c2, w2 := newImageCtxWithRequest(httptest.NewRequest(http.MethodGet, "/token", nil))
	api.GenUpToken(c2)
	if code := readCode2(t, w2); code != 0 {
		t.Fatalf("七牛启用后应成功, body=%s", w2.Body.String())
	}
}
