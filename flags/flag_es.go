package flags

import (
	"myblogx/models"
	"myblogx/service/es_service"
)

func FlagESIndex() {
	// 初始化ES索引
	article := models.ArticleModel{}
	es_service.CreateIndexForce(article.Index(), article.Mapping())
}
