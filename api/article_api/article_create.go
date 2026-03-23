package article_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/es_service"
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

	safeContent := markdown.MdToSafe(cr.Content)
	if cr.Abstract == "" {
		textContent := markdown.MdToText(safeContent)
		cr.Abstract = markdown.ExtractText(textContent, 200)
	}

	article := models.ArticleModel{
		AuthorID:       claims.UserID,
		Title:          cr.Title,
		Abstract:       cr.Abstract,
		Content:        safeContent,
		CategoryID:     cr.CategoryID,
		Cover:          cr.Cover,
		CommentsToggle: cr.CommentsToggle,
		Status:         cr.Status,
	}

	if global.Config.Site.Article.SkipExamining && cr.Status == enum.ArticleStatusExamining {
		article.Status = enum.ArticleStatusPublished
	}

	tagIDs := extractTagIDs(tagList)
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&article).Error; err != nil {
			return err
		}
		return syncArticleTags(tx, article.ID, tagIDs)
	}); err != nil {
		res.FailWithMsg("创建文章失败", c)
		return
	}

	applyTagArticleCountDelta(buildTagArticleCountDelta(nil, tagIDs))
	if len(tagIDs) > 0 {
		if err := es_service.UpdateESDocsTags([]uint{article.ID}); err != nil {
			global.Logger.Errorf("创建文章后刷新 ES 标签失败 article_id=%d err=%v", article.ID, err)
		}
	}
	res.OkWithMsg("创建文章成功", c)
}
