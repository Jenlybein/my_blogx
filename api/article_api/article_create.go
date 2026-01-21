package article_api

import (
	"bytes"
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/utils/jwts"
	"myblogx/utils/markdown"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

type ArticleCreateRequest struct {
	Title          string             `json:"title" binding:"required"`
	Abstract       string             `json:"abstract"`
	Content        string             `json:"content" binding:"required"`
	CategoryID     uint               `json:"category_id"`
	TagList        ctype.List         `json:"tag_list"`
	Cover          string             `json:"cover"`
	CommentsToggle bool               `json:"comments_toggle"`
	Status         enum.ArticleStatus `json:"status" binding:"required,oneof=1 2"`
}

func (ArticleApi) ArticleCreateView(c *gin.Context) {
	cr := middleware.GetBindJson[ArticleCreateRequest](c)

	claims := jwts.GetClaimsByGin(c)
	if err := global.DB.Take(&models.UserModel{}, claims.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	// 判断分类id是否存在
	var category models.CategoryModel
	if err := global.DB.Take(&category, "id = ? AND user_id = ?", cr.CategoryID, claims.UserID).Error; err != nil {
		res.FailWithMsg("分类不存在", c)
		return
	}

	// 文章正文防止 xss 注入
	contentDoc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(cr.Content)))
	if err != nil {
		res.FailWithMsg("解析正文失败", c)
		return
	}
	contentDoc.Find("img").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("src")
		if exists {
			alt := s.Text()
			if alt == "" {
				alt = "image"
			}
			// 替换 img 标签为 Markdown 图片语法
			s.ReplaceWithHtml(fmt.Sprintf("![%s](%s)", alt, url))
		}
	})
	contentDoc.Find("script, style, iframe, embed, object, param, video, audio, source, track, menu").Remove()
	cr.Content, _ = contentDoc.Html()

	// 不传简介，则从正文中提取前 200 个字符
	if cr.Abstract == "" {
		// 把正文中的 markdown 格式去掉
		html := markdown.MdToHTML(cr.Content)
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(html)))
		if err != nil {
			res.FailWithMsg("解析正文失败", c)
			return
		}
		htmlText := doc.Text()
		if len(htmlText) > 200 {
			cr.Abstract = htmlText[:200]
		} else {
			cr.Abstract = htmlText
		}
	}

	// 正文内容图片转存

	var article = models.ArticleModel{
		AuthorID:       claims.UserID,
		Title:          cr.Title,
		Abstract:       cr.Abstract,
		Content:        cr.Content,
		CategoryID:     &cr.CategoryID,
		TagList:        cr.TagList,
		Cover:          &cr.Cover,
		CommentsToggle: cr.CommentsToggle,
		Status:         enum.ArticleStatusDraft,
	}

	if global.Config.Site.Article.SkipExamining {
		article.Status = enum.ArticleStatusPublished
	}

	if err := global.DB.Create(&article).Error; err != nil {
		res.FailWithMsg("创建文章失败", c)
		return
	}

	// 可以返回文章id
	res.OkWithMsg("创建文章成功", c)
}

// func xssFilter(content string) (string, error) {
// 	contentDoc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(content)))
// 	if err != nil {
// 		return content, errors.New("解析正文失败")
// 	}

// 	// 提取img标签的图片url，转为markdown格式
// 	contentDoc.Find("img").Map(func(i int, s *goquery.Selection) string {
// 		url, _ := s.Attr("src")
// 		return fmt.Sprintf("![%s](%s)", s.Text(), url)
// 	})

// 	// 过滤不需要的标签
// 	contentDoc.Find("script, style, iframe, embed, object, param, video, audio, source, track, menu").Remove()

// 	filteredContent := contentDoc.Text()

// 	return filteredContent, nil
// }
