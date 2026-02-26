package markdown_test

import (
	"myblogx/utils/markdown"
	"strings"
	"testing"
)

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
