package cron_service

import (
	"myblogx/global"
	"myblogx/models"
	"myblogx/service/redis_service/redis_article"
)

func SyncArticle() {
	favoriteMap := redis_article.GetAllCacheFavorite()
	diggMap := redis_article.GetAllCacheDigg()
	viewMap := redis_article.GetAllCacheView()

	// 获取所有缓存内的文章id
	articleIDs := make([]uint, 0, len(favoriteMap)+len(diggMap)+len(viewMap))
	for id := range favoriteMap {
		articleIDs = append(articleIDs, id)
	}
	for id := range diggMap {
		articleIDs = append(articleIDs, id)
	}
	for id := range viewMap {
		articleIDs = append(articleIDs, id)
	}

	var list []models.ArticleModel
	global.DB.Find(&list)

	for _, model := range list {
		favorite := favoriteMap[model.ID]
		digg := diggMap[model.ID]
		view := viewMap[model.ID]
		if favorite != 0 {
			model.FavorCount = favorite
		}
		if digg != 0 {
			model.DiggCount = digg
		}
		if view != 0 {
			model.ViewCount = view
		}
		if err := global.DB.Save(&model).Err(); err != nil {
			global.Logger.Errorf("同步任务，文章数据更新出错 err: %v", err)
			continue
		}

		global.Logger.Infof("同步任务，文章数据更新成功 id: %d", model.ID)
	}

	// 同步过程中可能存在新数据，获取一次
	favoriteMap = redis_article.GetAllCacheFavorite()
	diggMap = redis_article.GetAllCacheDigg()
	viewMap = redis_article.GetAllCacheView()

	// 清空当前文章数据缓存
	redis_article.ClearAllCacheArticle()

	// 更新缓存数据
	for id, favorite := range favoriteMap {
		redis_article.SetCacheFavorite(id, favorite)
	}
	for id, digg := range diggMap {
		redis_article.SetCacheDigg(id, digg)
	}
	for id, view := range viewMap {
		redis_article.SetCacheView(id, view)
	}
}
