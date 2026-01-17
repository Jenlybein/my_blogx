package flag_service

import (
	"myblogx/global"
	"myblogx/models"
	"myblogx/service/es_service"
)

func FlagESIndex() {
	// 初始化ES索引
	article := models.ArticleModel{}
	es_service.CreateIndexForce(global.Config.ES.Index, article.Mapping())
}
