package article_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/utils/jwts"
	"myblogx/utils/markdown"

	"github.com/gin-gonic/gin"
)

type ArticleCreateRequest struct {
	Title          string             `json:"title" binding:"required"`
	Abstract       string             `json:"abstract"`
	Content        string             `json:"content" binding:"required"`
	CategoryID     *uint              `json:"category_id"`
	TagList        ctype.List         `json:"tag_list"`
	Cover          string             `json:"cover"`
	CommentsToggle bool               `json:"comments_toggle"`
	Status         enum.ArticleStatus `json:"status" binding:"required,oneof=1 2"`
}

func (ArticleApi) ArticleCreateView(c *gin.Context) {
	cr := middleware.GetBindJson[ArticleCreateRequest](c)

	claims := jwts.MustGetClaimsByGin(c)
	if err := global.DB.Take(&models.UserModel{}, claims.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	// 判断分类id是否存在
	if cr.CategoryID != nil {
		var category models.CategoryModel
		if err := global.DB.Take(&category, "id = ? AND user_id = ?", *cr.CategoryID, claims.UserID).Error; err != nil {
			res.FailWithMsg("分类不存在", c)
			return
		}
	}
	// 从正文中提取纯文本内容
	textContent := markdown.MdToText(cr.Content)

	// 文章正文防止 xss 注入，安全转为 html 格式
	htmlContent := markdown.MdToHTMLSafe(cr.Content)

	// 不传简介，则从正文中提取前 200 个字符
	if cr.Abstract == "" {
		if len(textContent) > 200 {
			// 把字符串转为rune切片（rune对应单个Unicode字符）
			runes := []rune(textContent)
			// 安全截取子字符串
			cr.Abstract = string(runes[:200])
		} else {
			cr.Abstract = textContent
		}
	}

	var article = models.ArticleModel{
		AuthorID:       claims.UserID,
		Title:          cr.Title,
		Abstract:       cr.Abstract,
		Content:        cr.Content,
		HtmlContent:    htmlContent,
		CategoryID:     cr.CategoryID,
		TagList:        cr.TagList,
		Cover:          cr.Cover,
		CommentsToggle: cr.CommentsToggle,
		Status:         cr.Status,
	}

	if global.Config.Site.Article.SkipExamining {
		if cr.Status == enum.ArticleStatusExamining {
			article.Status = enum.ArticleStatusPublished
		}
		// TODO：审核 textContent 内容，判断是否包含违规内容
	}

	if err := global.DB.Create(&article).Error; err != nil {
		res.FailWithMsg("创建文章失败", c)
		return
	}

	// 可以返回文章id
	res.OkWithMsg("创建文章成功", c)
}
