package markdown

import (
	"bytes"
	"regexp"

	mdrenderer "github.com/blackstork-io/goldmark-markdown"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

var (
	// 图片缩放正则表达式
	reZoomValue = regexp.MustCompile(`^((0\.\d+|1(\.\d+)?|2(\.0+)?)|([1-9]\d?%|1\d{2}%|200%))$`)
	// 匹配 math, math inline, math display
	reMath = regexp.MustCompile(`^math(\s.+)?$`)
	// 匹配 user-content- 开头的 id
	reUserContentID = regexp.MustCompile(`^user-content-.*$`)
)

var rawHTMLMarkdownEngine = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.DefinitionList,
		mathjax.MathJax,
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithUnsafe(),
	),
)

var safeMarkdownEngine = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.DefinitionList,
		mathjax.MathJax,
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
	),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRenderer(mdrenderer.NewRenderer()),
)

func getRawHTML(md string) string {
	source := []byte(md)
	doc := rawHTMLMarkdownEngine.Parser().Parse(text.NewReader(source))

	var buf bytes.Buffer
	if err := rawHTMLMarkdownEngine.Renderer().Render(&buf, source, doc); err != nil {
		return ""
	}
	return buf.String()
}
