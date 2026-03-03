package redis_comment

import (
	"context"
	"fmt"
	"myblogx/global"
	"strconv"
)

const ReplyCountCacheKey = "comment_reply"

func SetCacheReply(commentID uint, increase int) error {
	return global.Redis.HIncrBy(context.Background(), ReplyCountCacheKey, strconv.Itoa(int(commentID)), int64(increase)).Err()
}

func GetCacheReply(commentID uint) int {
	num, _ := global.Redis.HGet(context.Background(), ReplyCountCacheKey, strconv.Itoa(int(commentID))).Int()
	return num
}

func GetBatchCacheReply(commentIDs []uint) map[uint]int {
	result := make(map[uint]int, len(commentIDs))
	if len(commentIDs) == 0 {
		return result
	}

	fields := make([]string, 0, len(commentIDs))
	for _, commentID := range commentIDs {
		fields = append(fields, strconv.Itoa(int(commentID)))
	}

	values, err := global.Redis.HMGet(context.Background(), ReplyCountCacheKey, fields...).Result()
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
		result[commentIDs[i]] = num
	}
	return result
}

func GetAllCacheReply() map[uint]int {
	res, err := global.Redis.HGetAll(context.Background(), ReplyCountCacheKey).Result()
	if err != nil {
		return nil
	}

	numMap := make(map[uint]int, len(res))
	for k, v := range res {
		commentID, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			continue
		}
		num, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		numMap[uint(commentID)] = num
	}
	return numMap
}

func ClearAllCacheReply() error {
	return global.Redis.Del(context.Background(), ReplyCountCacheKey).Err()
}
