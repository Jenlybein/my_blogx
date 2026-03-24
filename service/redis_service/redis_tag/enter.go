package redis_tag

import (
	"context"
	"fmt"
	"myblogx/models/ctype"
	"strconv"

	"myblogx/global"
)

const TagCacheArticleCount = "tag_article_count"

func SetCacheArticleCount(tagID ctype.ID, increase int) error {
	if global.Redis == nil {
		return nil
	}
	return global.Redis.HIncrBy(context.Background(), TagCacheArticleCount, tagID.String(), int64(increase)).Err()
}

func GetCacheArticleCount(tagID ctype.ID) int {
	if global.Redis == nil {
		return 0
	}
	num, _ := global.Redis.HGet(context.Background(), TagCacheArticleCount, tagID.String()).Int()
	return num
}

func GetBatchCacheArticleCount(tagIDs []ctype.ID) map[ctype.ID]int {
	result := make(map[ctype.ID]int, len(tagIDs))
	if len(tagIDs) == 0 {
		return result
	}
	if global.Redis == nil {
		return result
	}

	fields := make([]string, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		fields = append(fields, tagID.String())
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

func GetAllCacheArticleCount() map[ctype.ID]int {
	if global.Redis == nil {
		return map[ctype.ID]int{}
	}
	res, err := global.Redis.HGetAll(context.Background(), TagCacheArticleCount).Result()
	if err != nil {
		return nil
	}

	numMap := make(map[ctype.ID]int, len(res))
	for k, v := range res {
		tagID, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		num, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		numMap[ctype.ID(tagID)] = num
	}
	return numMap
}

func ClearAllCacheTag() error {
	if global.Redis == nil {
		return nil
	}
	return global.Redis.Del(context.Background(), TagCacheArticleCount).Err()
}
