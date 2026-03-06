package redis_tag

import (
	"context"
	"fmt"
	"strconv"

	"myblogx/global"
)

const TagCacheArticleCount = "tag_article_count"

func SetCacheArticleCount(tagID uint, increase int) error {
	if global.Redis == nil {
		return nil
	}
	return global.Redis.HIncrBy(context.Background(), TagCacheArticleCount, strconv.Itoa(int(tagID)), int64(increase)).Err()
}

func GetCacheArticleCount(tagID uint) int {
	if global.Redis == nil {
		return 0
	}
	num, _ := global.Redis.HGet(context.Background(), TagCacheArticleCount, strconv.Itoa(int(tagID))).Int()
	return num
}

func GetBatchCacheArticleCount(tagIDs []uint) map[uint]int {
	result := make(map[uint]int, len(tagIDs))
	if len(tagIDs) == 0 {
		return result
	}
	if global.Redis == nil {
		return result
	}

	fields := make([]string, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		fields = append(fields, strconv.Itoa(int(tagID)))
	}

	values, err := global.Redis.HMGet(context.Background(), TagCacheArticleCount, fields...).Result()
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
		result[tagIDs[i]] = num
	}
	return result
}

func GetAllCacheArticleCount() map[uint]int {
	if global.Redis == nil {
		return map[uint]int{}
	}
	res, err := global.Redis.HGetAll(context.Background(), TagCacheArticleCount).Result()
	if err != nil {
		return nil
	}

	numMap := make(map[uint]int, len(res))
	for k, v := range res {
		tagID, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		num, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		numMap[uint(tagID)] = num
	}
	return numMap
}

func ClearAllCacheTag() error {
	if global.Redis == nil {
		return nil
	}
	return global.Redis.Del(context.Background(), TagCacheArticleCount).Err()
}
