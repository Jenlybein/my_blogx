package article_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_article"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (ArticleApi) ArticleDetailView(c *gin.Context) {
	cr := middleware.GetBindUri[models.IDRequest](c)

	var article models.ArticleModel
	if err := global.DB.Select(
		"ID",
		"CreatedAt",
		"UpdatedAt",
		"Title",
		"Abstract",
		"Content",
		"Cover",
		"AuthorID",
		"ViewCount",
		"DiggCount",
		"CommentCount",
		"FavorCount",
		"CommentsToggle",
		"Status",
	).Preload("UserModel").
		Preload("CategoryModel").
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort desc, id asc")
		}).
		Take(&article, cr.ID).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}

	claims, _ := jwts.ParseTokenByGin(c)
	if claims == nil {
		if article.Status != enum.ArticleStatusPublished {
			res.FailWithMsg("文章不存在", c)
			return
		}
	} else if claims.Role == enum.RoleUser && article.AuthorID != claims.UserID && article.Status != enum.ArticleStatusPublished {
		res.FailWithMsg("文章不存在", c)
		return
	}

	counters := redis_article.GetBatchCounters([]uint{article.ID})
	article.DiggCount += counters.DiggMap[article.ID]
	article.ViewCount += counters.ViewMap[article.ID]
	article.FavorCount += counters.FavorMap[article.ID]
	article.CommentCount += counters.CommentMap[article.ID]

	response := ArticleDetailResponse{
		ID:             article.ID,
		CreatedAt:      article.CreatedAt,
		UpdatedAt:      article.UpdatedAt,
		Title:          article.Title,
		Abstract:       article.Abstract,
		Content:        article.Content,
		Cover:          article.Cover,
		ViewCount:      article.ViewCount,
		DiggCount:      article.DiggCount,
		CommentCount:   article.CommentCount,
		FavorCount:     article.FavorCount,
		CommentsToggle: article.CommentsToggle,
		Status:         article.Status,
		AuthorAvatar:   article.UserModel.Avatar,
		AuthorNickname: article.UserModel.Nickname,
		AuthorUsername: article.UserModel.Username,
	}
	if article.CategoryModel != nil {
		response.CategoryName = article.CategoryModel.Title
	}
	if article.Tags != nil {
		for _, tag := range article.Tags {
			response.Tags = append(response.Tags, tag.Title)
		}
	}

	res.OkWithData(response, c)
}
