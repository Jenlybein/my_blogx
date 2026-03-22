package article_api

import (
	"myblogx/api/article_api/category"
	"myblogx/api/article_api/favorite"
	"myblogx/api/article_api/tags"
	"myblogx/api/article_api/top"
	"myblogx/api/article_api/view_history"
)

type ArticleApi struct {
	category.CategoryApi
	favorite.FavoriteApi
	top.TopApi
	view_history.ViewHistoryApi
	tags.TagsApi
}
