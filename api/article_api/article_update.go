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

	updateMap := map[string]any{}

	if cr.Title != nil {
		updateMap["title"] = *cr.Title
	}
	if cr.Content != nil {
		updateMap["content"] = markdown.MdToSafe(*cr.Content)
	}
	if cr.Abstract != nil {
		abstract := *cr.Abstract
		if abstract == "" {
			content := article.Content
			if cr.Content != nil {
				content = markdown.MdToSafe(*cr.Content)
			}
			textContent := markdown.MdToText(content)
			abstract = markdown.ExtractText(textContent, 200)
		}
		updateMap["abstract"] = abstract
	}

	if cr.CategoryID != nil {
		if *cr.CategoryID == 0 {
			updateMap["category_id"] = nil
		} else {
			if err := validateArticleCategory(global.DB, claims.UserID, cr.CategoryID); err != nil {
				res.FailWithMsg("分类不存在", c)
				return
			}
			updateMap["category_id"] = cr.CategoryID
		}
	}
	if cr.Cover != nil {
		updateMap["cover"] = *cr.Cover
	}
	if cr.CommentsToggle != nil {
		updateMap["comments_toggle"] = *cr.CommentsToggle
	}

	var (
		tagList   []models.TagModel
		oldTagIDs []uint
		newTagIDs []uint
		err       error
	)
	if cr.TagIDs != nil {
		oldTagIDs, err = loadArticleTagIDs(global.DB, article.ID)
		if err != nil {
			res.FailWithMsg("查询文章标签失败", c)
			return
		}

		tagList, err = loadEnabledTagsByIDs(global.DB, *cr.TagIDs)
		if err != nil {
			res.FailWithMsg(err.Error(), c)
			return
		}
		newTagIDs = extractTagIDs(tagList)
	}

	if len(updateMap) == 0 {
		res.OkWithMsg("更新文章成功", c)
		return
	}

	if !global.Config.Site.Article.SkipExamining && article.Status == enum.ArticleStatusPublished {
		updateMap["status"] = enum.ArticleStatusExamining
	}

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&article).Updates(updateMap).Error; err != nil {
			return err
		}
		if cr.TagIDs != nil {
			return tx.Model(&article).Association("Tags").Replace(tagList)
		}
		return nil
	}); err != nil {
		res.FailWithMsg("更新文章失败", c)
		return
	}

	if cr.TagIDs != nil {
		applyTagArticleCountDelta(buildTagArticleCountDelta(oldTagIDs, newTagIDs))
		if err := es_service.UpdateESDocsTags([]uint{article.ID}); err != nil {
			global.Logger.Errorf("更新文章标签后刷新 ES 标签失败 article_id=%d err=%v", article.ID, err)
		}
	} else if cr.Content != nil {
		if err := es_service.UpdateESDocsContent([]uint{article.ID}); err != nil {
			global.Logger.Errorf("更新文章正文后刷新 ES 文档失败 article_id=%d err=%v", article.ID, err)
		}
	}
	res.OkWithMsg("更新文章成功", c)
}
