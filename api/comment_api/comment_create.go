package comment_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/service/redis_service/redis_article"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentCreateRequest struct {
	Content   string `json:"content" binding:"required"`
	ArticleID uint   `json:"article_id" binding:"required"`
	ParentID  *uint  `json:"parent_id"` // 父级评论id
}

func (CommentApi) CommentCreateView(c *gin.Context) {
	cr := middleware.GetBindJson[CommentCreateRequest](c)

	var article models.ArticleModel
	if err := global.DB.Take(&article, cr.ArticleID).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}
	if !article.CommentsToggle {
		res.FailWithMsg("该文章已关闭评论", c)
		return
	}

	claims := jwts.MustGetClaimsByGin(c)
	model := models.CommentModel{
		Content:   cr.Content,
		UserID:    claims.UserID,
		ArticleID: cr.ArticleID,
	}

	if cr.ParentID != nil {
		rootParentID, err := findRootParentID(cr.ArticleID, *cr.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				res.FailWithMsg("父评论不存在", c)
				return
			}
			res.FailWithMsg("查询父评论失败", c)
			return
		}
		model.ParentID = cr.ParentID
		model.RootParentID = rootParentID
	}

	if err := global.DB.Create(&model).Error; err != nil {
		res.FailWithMsg("评论失败", c)
		return
	}
	if err := redis_article.SetCacheComment(cr.ArticleID, 1); err != nil {
		global.Logger.Errorf("写入评论计数缓存失败 article_id=%d err=%v", cr.ArticleID, err)
	}

	res.OkWithMsg("评论成功", c)
}

// 找到父评论所属的根评论ID。
// 一级评论的回复：根评论=父评论ID；回复楼中楼：根评论沿用父评论的 RootParentID。
func findRootParentID(articleID, parentID uint) (*uint, error) {
	var parent models.CommentModel
	if err := global.DB.Take(&parent, "id = ? and article_id = ?", parentID, articleID).Error; err != nil {
		return nil, err
	}
	if parent.RootParentID != nil {
		return parent.RootParentID, nil
	}
	return &parent.ID, nil
}
