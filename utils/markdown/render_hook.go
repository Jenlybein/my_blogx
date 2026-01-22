package markdown

import (
	"io"
	"net/url"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

const AnchorPrefix = "user-content-"

// 锚点前缀钩子
func PrefixAnchorHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	// 安全处理标题节点 (锚点定义处)
	if heading, ok := node.(*ast.Heading); ok && entering {
		// 初始化
		if heading.Attribute == nil {
			heading.Attribute = &ast.Attribute{}
		}

		// 如果 ID 为空，生成对应 ID
		if len(heading.Attribute.ID) == 0 {
			var content []string
			ast.WalkFunc(heading, func(node ast.Node, entering bool) ast.WalkStatus {
				if text, ok := node.(*ast.Text); ok && entering {
					content = append(content, string(text.Leaf.Literal))
				}
				return ast.GoToNext
			})

			rawID := strings.Join(content, "")
			rawID = strings.TrimSpace(rawID)
			heading.Attribute.ID = []byte(rawID)
		}

		if len(heading.Attribute.ID) > 0 {
			// 对 ID 进行 URL 编码
			unescaped, _ := url.QueryUnescape(string(heading.Attribute.ID))
			encodedID := url.PathEscape(unescaped)
			// 给 ID 加上前缀
			newID := AnchorPrefix + encodedID
			heading.Attribute.ID = []byte(newID)
		}
	}

	// 安全处理链接节点 (锚点跳转处)
	if link, ok := node.(*ast.Link); ok && entering {
		dest := string(link.Destination)
		// 如果是本页锚点链接（以 # 开头）
		if strings.HasPrefix(dest, "#") {
			rawContent := strings.TrimLeft(dest, "#")
			rawContent = strings.TrimSpace(rawContent)
			if len(rawContent) > 0 {
				// 对 ID 内容进行解码再编码（防止双重编码）
				unescaped, _ := url.QueryUnescape(rawContent)
				encodedID := url.PathEscape(unescaped)

				// 强制只保留一个 # 并加上前缀
				newDest := "#" + AnchorPrefix + encodedID
				link.Destination = []byte(newDest)
			}
		}
	}

	return ast.GoToNext, false
}
