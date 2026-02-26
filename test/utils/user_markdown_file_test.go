package utils_test

import (
	"bytes"
	"encoding/base64"
	"mime/multipart"
	"myblogx/test/testutil"
	"myblogx/utils/file"
	"myblogx/utils/markdown"
	"myblogx/utils/user_info"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserInfoHelpers(t *testing.T) {
	if user_info.IpType("127.0.0.1") != "ipv4" {
		t.Fatal("ipv4 识别失败")
	}
	if user_info.IpType("::1") != "ipv6" {
		t.Fatal("ipv6 识别失败")
	}
	if user_info.IpType("not-ip") != "" {
		t.Fatal("非法 IP 应返回空")
	}

	if !user_info.IsLocalIP("127.0.0.1", "ipv4") {
		t.Fatal("127.0.0.1 应被识别为本地 IP")
	}
	if !user_info.IsLocalIP("192.168.1.2", "ipv4") {
		t.Fatal("192.168.x.x 应被识别为本地 IP")
	}
	if user_info.IsLocalIP("8.8.8.8", "ipv4") {
		t.Fatal("公网 IP 不应被识别为本地 IP")
	}
}

func TestGetClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("X-Forwarded-For", "10.0.0.2, 8.8.8.8")
	c.Request = req

	if ip := user_info.GetClientIP(c); ip != "10.0.0.2" {
		t.Fatalf("X-Forwarded-For 提取失败: %s", ip)
	}
}

func TestMarkdownHelpers(t *testing.T) {
	md := "# 标题\n\n<script>alert(1)</script>\n\n[跳转](#标题)"
	unsafe := markdown.MdToHTMLUnsafe(md)
	if !strings.Contains(unsafe, "<script>") {
		t.Fatal("MdToHTMLUnsafe 不应过滤 script")
	}
	if !strings.Contains(unsafe, "user-content-") {
		t.Fatal("锚点前缀未生效")
	}

	safe := markdown.MdToHTMLSafe(md)
	if strings.Contains(safe, "<script>") {
		t.Fatal("MdToHTMLSafe 应过滤 script")
	}
	if !strings.Contains(safe, "user-content-") {
		t.Fatal("安全模式仍应保留锚点前缀")
	}

	text := markdown.MdToText("**hello**")
	if strings.Contains(text, "<") {
		t.Fatalf("MdToText 应返回纯文本: %s", text)
	}

	if got := markdown.ExtractText("abcdef", 3); got != "abc" {
		t.Fatalf("ExtractText 截断错误: %s", got)
	}
}

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
