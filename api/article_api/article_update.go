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

type ArticleUpdateRequest struct {
	Title          string     `json:"title" binding:"required"`
	Abstract       string     `json:"abstract"`
	Content        string     `json:"content" binding:"required"`
	CategoryID     *uint      `json:"category_id"`
	TagList        ctype.List `json:"tag_list"`
	Cover          string     `json:"cover"`
	CommentsToggle bool       `json:"comments_toggle"`
}

func (ArticleApi) ArticleUpdateView(c *gin.Context) {
	id := middleware.GetBindUri[models.IDRequest](c)
	cr := middleware.GetBindJson[ArticleUpdateRequest](c)

	claims := jwts.MustGetClaimsByGin(c)
	if err := global.DB.Take(&models.UserModel{}, claims.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	// 查询文章是否存在
	var article models.ArticleModel
	if err := global.DB.Take(&article, "id = ? AND author_id = ?", id.ID, claims.UserID).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
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
	// 文章正文防止 xss 注入，安全转为 html 格式
	htmlContent := markdown.MdToHTMLSafe(cr.Content)

	// 不传简介，则从正文中提取前 200 个字符
	if cr.Abstract == "" {
		textContent := markdown.MdToText(cr.Content)
		cr.Abstract = markdown.ExtractText(textContent, 200)
	}

	updateMap := map[string]any{
		"title":           cr.Title,
		"abstract":        cr.Abstract,
		"content":         cr.Content,
		"html_content":    htmlContent,
		"category_id":     cr.CategoryID,
		"tag_list":        cr.TagList,
		"cover":           cr.Cover,
		"comments_toggle": cr.CommentsToggle,
	}

	if !global.Config.Site.Article.SkipExamining && article.Status == enum.ArticleStatusPublished {
		updateMap["status"] = enum.ArticleStatusExamining
		// TODO：审核 textContent 内容，判断是否包含违规内容
	}

	if err := global.DB.Model(&article).Updates(updateMap).Error; err != nil {
		res.FailWithMsg("更新文章失败", c)
		return
	}

	// 可以返回文章id
	res.OkWithMsg("更新文章成功", c)
}
