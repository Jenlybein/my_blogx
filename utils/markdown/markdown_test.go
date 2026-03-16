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
	if !strings.Contains(unsafe, `href="#`) {
		t.Fatal("Markdown 链接应被渲染为 HTML 链接")
	}

	safe := markdown.MdToHTMLSafe(md)
	if strings.Contains(safe, "<script>") {
		t.Fatal("MdToHTMLSafe 应过滤 script")
	}
	if !strings.Contains(safe, `href="#`) {
		t.Fatal("安全模式应保留普通锚点链接")
	}

	text := markdown.MdToText("**hello**")
	if strings.Contains(text, "<") {
		t.Fatalf("MdToText 应返回纯文本: %s", text)
	}

	if got := markdown.ExtractText("abcdef", 3); got != "abc" {
		t.Fatalf("ExtractText 截断错误: %s", got)
	}
}

func TestMdToSafe(t *testing.T) {
	md := "# 标题\n\n<script>alert(1)</script>\n\n[跳转](#标题)\n\n[危险](javascript:alert(1))\n\n![图](javascript:alert(2))"
	safe := markdown.MdToSafe(md)

	if strings.Contains(safe, "<script>") {
		t.Fatalf("MdToSafe 应移除原生 HTML: %s", safe)
	}
	if strings.Contains(safe, "javascript:") {
		t.Fatalf("MdToSafe 应移除危险链接: %s", safe)
	}
	if !strings.Contains(safe, "[跳转](#标题)") {
		t.Fatalf("MdToSafe 应保留原始锚点链接: %s", safe)
	}
}

func TestMdToSafeKeepsAllowedHTML(t *testing.T) {
	md := `<span class="math inline">x^2</span>

<img src="https://example.com/a.png" style="zoom:50%">`

	safe := markdown.MdToSafe(md)
	if !strings.Contains(safe, `<span>x^2</span>`) {
		t.Fatalf("MdToSafe 应保留安全 HTML 内容: %s", safe)
	}
	if strings.Contains(safe, `class="math inline"`) {
		t.Fatalf("MdToSafe 不应保留前端渲染不需要的 math class: %s", safe)
	}
	if !strings.Contains(safe, `style="zoom: 50%"`) {
		t.Fatalf("MdToSafe 应保留图片 zoom 样式: %s", safe)
	}
}

func TestMdToHTMLSafeKeepsMathClass(t *testing.T) {
	md := `<span class="math inline">x^2</span>`
	safe := markdown.MdToHTMLSafe(md)

	if !strings.Contains(safe, `<span class="math inline">x^2</span>`) {
		t.Fatalf("MdToHTMLSafe 应保留 math class: %s", safe)
	}
}

func TestMdToSafeKeepsMathSyntax(t *testing.T) {
	md := "行内公式 $a+b$\n\n$$\nx^2+y^2=z^2\n$$"
	safe := markdown.MdToSafe(md)

	if !strings.Contains(safe, "$a+b$") {
		t.Fatalf("MdToSafe 应保留行内数学公式: %s", safe)
	}
	if !strings.Contains(safe, "$$") || !strings.Contains(safe, "x^2+y^2=z^2") {
		t.Fatalf("MdToSafe 应保留块级数学公式: %s", safe)
	}
}

func TestMdToSafeDropsMalformedURL(t *testing.T) {
	md := "[异常链接](http://[::1)"
	safe := markdown.MdToSafe(md)

	if strings.Contains(safe, "http://[::1") {
		t.Fatalf("MdToSafe 不应保留解析失败的链接: %s", safe)
	}
}
