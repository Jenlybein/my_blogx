package redis_article

import (
	"context"
	"myblogx/global"
	"strconv"
)

type articleCacheType string

const (
	articleCacheView     articleCacheType = "article_view"
	articleCacheDigg     articleCacheType = "article_digg"
	articleCacheFavorite articleCacheType = "article_favorite"
)

func set(t articleCacheType, articleID uint, increase int) error {
	num, _ := global.Redis.HGet(context.Background(), string(t), strconv.Itoa(int(articleID))).Int()
	num += increase
	return global.Redis.HSet(context.Background(), string(t), strconv.Itoa(int(articleID)), num).Err()
}

func SetCacheView(articleID uint, increase int) error {
	return set(articleCacheView, articleID, increase)
}
func SetCacheDigg(articleID uint, increase int) error {
	return set(articleCacheDigg, articleID, increase)
}
func SetCacheFavorite(articleID uint, increase int) error {
	return set(articleCacheFavorite, articleID, increase)
}
