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

type ArticleDetailResponse struct {
	models.ArticleModel
	AuthorAvatar   string `json:"author_avatar"`
	AuthorNickname string `json:"author_name"`
	AuthorUsername string `json:"author_username"`
	CategoryName   string `json:"category_name"`
}

func (ArticleApi) ArticleDetailView(c *gin.Context) {
	cr := middleware.GetBindUri[models.IDRequest](c)

	var article models.ArticleModel
	if err := global.DB.Preload("UserModel").
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

	article.DiggCount += redis_article.GetCacheDigg(article.ID)
	article.ViewCount += redis_article.GetCacheView(article.ID)
	article.FavorCount += redis_article.GetCacheFavorite(article.ID)
	article.CommentCount += redis_article.GetCacheComment(article.ID)

	response := ArticleDetailResponse{
		ArticleModel:   article,
		AuthorAvatar:   article.UserModel.Avatar,
		AuthorNickname: article.UserModel.Nickname,
		AuthorUsername: article.UserModel.Username,
	}
	if article.CategoryModel != nil {
		response.CategoryName = article.CategoryModel.Title
	}

	res.OkWithData(response, c)
}
