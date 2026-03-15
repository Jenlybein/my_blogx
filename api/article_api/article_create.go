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

func (ArticleApi) ArticleCreateView(c *gin.Context) {
	cr := middleware.GetBindJson[ArticleCreateRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	if err := global.DB.Take(&models.UserModel{}, claims.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	if claims.Role != enum.RoleAdmin && global.Config.Site.SiteInfo.Mode == enum.SiteModeBlog {
		res.FailWithMsg("站点处于个人博客模式，普通用户无法创建文章", c)
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

	article := models.ArticleModel{
		AuthorID:       claims.UserID,
		Title:          cr.Title,
		Abstract:       cr.Abstract,
		Content:        cr.Content,
		HtmlContent:    htmlContent,
		CategoryID:     cr.CategoryID,
		Cover:          cr.Cover,
		CommentsToggle: cr.CommentsToggle,
		Status:         cr.Status,
		TagList:        extractTagTitles(tagList),
	}

	if global.Config.Site.Article.SkipExamining && cr.Status == enum.ArticleStatusExamining {
		article.Status = enum.ArticleStatusPublished
	}

	tagIDs := extractTagIDs(tagList)
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&article).Error; err != nil {
			return err
		}
		return tx.Model(&article).Association("Tags").Replace(tagList)
	}); err != nil {
		res.FailWithMsg("创建文章失败", c)
		return
	}

	applyTagArticleCountDelta(buildTagArticleCountDelta(nil, tagIDs))
	res.OkWithMsg("创建文章成功", c)
}
