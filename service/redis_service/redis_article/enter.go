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

const (
	ArticleCacheGuestView articleCacheType = "article_guest_view"
)

func set(t articleCacheType, articleID uint, increase int) error {
	// num, _ := global.Redis.HGet(context.Background(), string(t), strconv.Itoa(int(articleID))).Int()
	// num += increase
	// return global.Redis.HSet(context.Background(), string(t), strconv.Itoa(int(articleID)), num).Err()

	return global.Redis.HIncrBy(context.Background(), string(t), strconv.Itoa(int(articleID)), int64(increase)).Err()
}

func get(t articleCacheType, articleID uint) int {
	num, _ := global.Redis.HGet(context.Background(), string(t), strconv.Itoa(int(articleID))).Int()
	return num
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

func GetCacheView(articleID uint) int {
	return get(articleCacheView, articleID)
}
func GetCacheDigg(articleID uint) int {
	return get(articleCacheDigg, articleID)
}
func GetCacheFavorite(articleID uint) int {
	return get(articleCacheFavorite, articleID)
}

func GetAll(t articleCacheType) map[uint]int {
	res, err := global.Redis.HGetAll(context.Background(), string(t)).Result()
	if err != nil {
		return nil
	}
	numMap := make(map[uint]int)
	for k, v := range res {
		ik, err := strconv.Atoi(k)
		num, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		numMap[uint(ik)] = num
	}
	return numMap
}

func GetAllCacheView() map[uint]int {
	return GetAll(articleCacheView)
}
func GetAllCacheDigg() map[uint]int {
	return GetAll(articleCacheDigg)
}
func GetAllCacheFavorite() map[uint]int {
	return GetAll(articleCacheFavorite)
}
