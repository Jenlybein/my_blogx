package redis_article

import (
	"context"
	"fmt"
	"myblogx/global"
	"strconv"
	"time"
)

type ArticleCacheType string

// 文章缓存的Key
const (
	ArticleCacheView     ArticleCacheType = "article_view"
	ArticleCacheDigg     ArticleCacheType = "article_digg"
	ArticleCacheFavorite ArticleCacheType = "article_favorite"
	ArticleCacheComment  ArticleCacheType = "article_comment"
)

// 设置缓存
func set(t ArticleCacheType, articleID uint, increase int) error {
	return global.Redis.HIncrBy(context.Background(), string(t), strconv.Itoa(int(articleID)), int64(increase)).Err()
}

func get(t ArticleCacheType, articleID uint) int {
	num, _ := global.Redis.HGet(context.Background(), string(t), strconv.Itoa(int(articleID))).Int()
	return num
}

// 浏览量缓存
func SetCacheView(articleID uint, increase int) error {
	return set(ArticleCacheView, articleID, increase)
}
func GetCacheView(articleID uint) int {
	return get(ArticleCacheView, articleID)
}

// 点赞缓存
func SetCacheDigg(articleID uint, increase int) error {
	return set(ArticleCacheDigg, articleID, increase)
}
func GetCacheDigg(articleID uint) int {
	return get(ArticleCacheDigg, articleID)
}

// 收藏缓存
func SetCacheFavorite(articleID uint, increase int) error {
	return set(ArticleCacheFavorite, articleID, increase)
}
func GetCacheFavorite(articleID uint) int {
	return get(ArticleCacheFavorite, articleID)
}

// 评论缓存
func SetCacheComment(articleID uint, increase int) error {
	return set(ArticleCacheComment, articleID, increase)
}
func GetCacheComment(articleID uint) int {
	return get(ArticleCacheComment, articleID)
}

func GetAll(t ArticleCacheType) map[uint]int {
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

func getBatch(t ArticleCacheType, articleIDs []uint) map[uint]int {
	result := make(map[uint]int, len(articleIDs))
	if len(articleIDs) == 0 {
		return result
	}

	fields := make([]string, 0, len(articleIDs))
	for _, articleID := range articleIDs {
		fields = append(fields, strconv.Itoa(int(articleID)))
	}

	values, err := global.Redis.HMGet(context.Background(), string(t), fields...).Result()
	if err != nil {
		return result
	}

	for i, raw := range values {
		if raw == nil {
			continue
		}
		num, err := strconv.Atoi(fmt.Sprint(raw))
		if err != nil {
			continue
		}
		result[articleIDs[i]] = num
	}
	return result
}

func GetBatchCacheView(articleIDs []uint) map[uint]int {
	return getBatch(ArticleCacheView, articleIDs)
}
func GetBatchCacheDigg(articleIDs []uint) map[uint]int {
	return getBatch(ArticleCacheDigg, articleIDs)
}
func GetBatchCacheFavorite(articleIDs []uint) map[uint]int {
	return getBatch(ArticleCacheFavorite, articleIDs)
}
func GetBatchCacheComment(articleIDs []uint) map[uint]int {
	return getBatch(ArticleCacheComment, articleIDs)
}

func GetAllCacheView() map[uint]int {
	return GetAll(ArticleCacheView)
}
func GetAllCacheDigg() map[uint]int {
	return GetAll(ArticleCacheDigg)
}
func GetAllCacheFavorite() map[uint]int {
	return GetAll(ArticleCacheFavorite)
}
func GetAllCacheComment() map[uint]int {
	return GetAll(ArticleCacheComment)
}

func ClearAllCacheArticle() error {
	return global.Redis.Del(
		context.Background(),
		string(ArticleCacheView),
		string(ArticleCacheDigg),
		string(ArticleCacheFavorite),
		string(ArticleCacheComment),
	).Err()
}

// 设置用户阅读历史
func SetUserArticleHistoryCache(articleID, userID int) {
	key := fmt.Sprintf("user_history_%d", userID)
	field := fmt.Sprintf("%d", articleID)

	now := time.Now()
	nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	if err := global.Redis.HSet(context.Background(), key, field, "").Err(); err != nil {
		global.Logger.Errorf("err: %v", err)
		return
	}

	if err := global.Redis.ExpireAt(context.Background(), key, nextDay).Err(); err != nil {
		global.Logger.Errorf("err: %v", err)
		return
	}
}
func GetUserArticleHistoryCache(articleID, userID int) bool {
	key := fmt.Sprintf("user_history_%d", userID)
	field := fmt.Sprintf("%d", articleID)

	_, err := global.Redis.HGet(context.Background(), key, field).Result()
	if err != nil {
		return false
	}
	return true
}

// 访客阅读记录
func SetGuestArticleHistoryCache(articleID int, hash string) {
	key := fmt.Sprintf("guest_history_%s", hash)
	field := fmt.Sprintf("%d", articleID)

	now := time.Now()
	nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	if err := global.Redis.HSet(context.Background(), key, field, "").Err(); err != nil {
		global.Logger.Errorf("err: %v", err)
		return
	}

	if err := global.Redis.ExpireAt(context.Background(), key, nextDay).Err(); err != nil {
		global.Logger.Errorf("err: %v", err)
		return
	}
}
func GetGuestArticleHistoryCache(articleID int, hash string) bool {
	key := fmt.Sprintf("guest_history_%s", hash)
	field := fmt.Sprintf("%d", articleID)

	_, err := global.Redis.HGet(context.Background(), key, field).Result()
	if err != nil {
		return false
	}
	return true
}
