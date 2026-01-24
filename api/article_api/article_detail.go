package article_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
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
	if err := global.DB.Preload("UserModel").Preload("CategoryModel").Take(&article, cr.ID).Error; err != nil {
		res.FailWithMsg("文章不存在", c)
		return
	}

	claims, _ := jwts.ParseTokenByGin(c)
	if claims == nil && article.Status != enum.ArticleStatusPublished {
		res.FailWithMsg("文章不存在", c)
		return
	}

	switch claims.Role {
	case enum.RoleUser:
		if article.AuthorID != claims.UserID {
			// 文章不是自己的
			if article.Status != enum.ArticleStatusPublished {
				res.FailWithMsg("文章不存在", c)
				return
			}
		}

	case enum.RoleAdmin:
		// 管理员可以查看所有文章
	}

	// TODO: 从缓存里面获取浏览量和点赞数

	var response = ArticleDetailResponse{
		ArticleModel:   article,
		AuthorAvatar:   article.UserModel.Avatar,
		AuthorNickname: article.UserModel.Nickname,
		AuthorUsername: article.UserModel.Username,
		CategoryName:   article.CategoryModel.Title,
	}

	res.OkWithData(response, c)
}
