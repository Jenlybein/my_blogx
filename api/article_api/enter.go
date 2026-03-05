package article_api

import (
	"myblogx/api/article_api/category"
	"myblogx/api/article_api/favorite"
	"myblogx/api/article_api/view_history"
)

type ArticleApi struct {
	category.CategoryApi
	favorite.FavoriteApi
	view_history.ViewHistoryApi
}
