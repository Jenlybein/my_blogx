package article_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/utils/jwts"
	"myblogx/utils/markdown"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleUpdateRequest struct {
	Title          string `json:"title" binding:"required"`
	Abstract       string `json:"abstract"`
	Content        string `json:"content" binding:"required"`
	CategoryID     *uint  `json:"category_id"`
	TagIDs         []uint `json:"tag_ids"`
	Cover          string `json:"cover"`
	CommentsToggle bool   `json:"comments_toggle"`
}

func (ArticleApi) ArticleUpdateView(c *gin.Context) {
	id := middleware.GetBindUri[models.IDRequest](c)
	cr := middleware.GetBindJson[ArticleUpdateRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	if err := global.DB.Take(&models.UserModel{}, claims.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	var article models.ArticleModel
	if err := global.DB.Take(&article, "id = ? AND author_id = ?", id.ID, claims.UserID).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}

	oldTagIDs, err := loadArticleTagIDs(global.DB, article.ID)
	if err != nil {
		res.FailWithMsg("查询文章标签失败", c)
		return
	}

	if err := validateArticleCategory(global.DB, claims.UserID, cr.CategoryID); err != nil {
		res.FailWithMsg("分类不存在", c)
		return
	}

	tagList, err := loadEnabledTagsByIDs(global.DB, cr.TagIDs)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	htmlContent := markdown.MdToHTMLSafe(cr.Content)
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
		"cover":           cr.Cover,
		"comments_toggle": cr.CommentsToggle,
	}

	if !global.Config.Site.Article.SkipExamining && article.Status == enum.ArticleStatusPublished {
		updateMap["status"] = enum.ArticleStatusExamining
	}

	newTagIDs := extractTagIDs(tagList)
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&article).Updates(updateMap).Error; err != nil {
			return err
		}
		return tx.Model(&article).Association("Tags").Replace(tagList)
	}); err != nil {
		res.FailWithMsg("更新文章失败", c)
		return
	}

	applyTagArticleCountDelta(buildTagArticleCountDelta(oldTagIDs, newTagIDs))
	res.OkWithMsg("更新文章成功", c)
}
