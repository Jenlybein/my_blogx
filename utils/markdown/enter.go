package markdown

import (
	"bytes"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark/text"
)

// 不允许任何 HTML 标签存在，纯文本输出
func MdToText(md string) string {
	raw := getRawHTML(md)
	// 移除所有 HTML 标签，只保留文本内容
	return bluemonday.StrictPolicy().Sanitize(raw)
}

// 允许所有 HTML 标签存在，不做任何过滤
func MdToHTMLUnsafe(md string) string {
	return getRawHTML(md)
}

// 过滤掉不安全的 HTML，如 <script>, <onerror> 等
func MdToHTMLSafe(md string) string {
	raw := getRawHTML(md)

	// 以 UGC 策略为基础（允许基本格式、链接等）
	// UGC策略已经很严格，想再禁用一些标签则使用SkipElementsContent
	p := bluemonday.UGCPolicy()

	// 允许数学公式用到的 class 和标签
	p.AllowAttrs("class").Matching(reMath).OnElements("span", "div")

	// 放行标题 id（用于标题锚点）
	p.AllowAttrs("id").Matching(reUserContentID).OnElements("h1", "h2", "h3", "h4", "h5", "h6")

	// 放行图片缩放（Typora 常用）
	p.AllowStyles("zoom").Matching(reZoomValue).OnElements("img")

	return p.Sanitize(raw)
}

// Markdown 内容过滤，确保安全合规
func MdToSafe(md string) string {
	source := []byte(md)
	doc := safeMarkdownEngine.Parser().Parse(text.NewReader(source))
	sanitizeSafeMarkdownAST(doc, source)

	var buf bytes.Buffer
	if err := safeMarkdownEngine.Renderer().Render(&buf, source, doc); err != nil {
		return ""
	}
	return buf.String()
}

// 提取纯文本前 n 个字符
func ExtractText(text string, length int) string {
	if len(text) > length {
		runes := []rune(text)
		return string(runes[:length])
	}
	return text
}
