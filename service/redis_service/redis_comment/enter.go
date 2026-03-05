package redis_comment

import (
	"context"
	"fmt"
	"myblogx/global"
	"strconv"
)

const ReplyCountCacheKey = "comment_reply"
const DiggCountCacheKey = "comment_digg"

func SetCacheReply(commentID uint, increase int) error {
	return set(ReplyCountCacheKey, commentID, increase)
}

func GetCacheReply(commentID uint) int {
	return get(ReplyCountCacheKey, commentID)
}

func DelCacheReply(commentID uint) error {
	return del(ReplyCountCacheKey, commentID)
}

func GetBatchCacheReply(commentIDs []uint) map[uint]int {
	return getBatch(ReplyCountCacheKey, commentIDs)
}

func GetAllCacheReply() map[uint]int {
	return getAll(ReplyCountCacheKey)
}

func ClearAllCacheReply() error {
	return global.Redis.Del(context.Background(), ReplyCountCacheKey).Err()
}

func SetCacheDigg(commentID uint, increase int) error {
	return set(DiggCountCacheKey, commentID, increase)
}

func GetCacheDigg(commentID uint) int {
	return get(DiggCountCacheKey, commentID)
}

func DelCacheDigg(commentID uint) error {
	return del(DiggCountCacheKey, commentID)
}

func GetBatchCacheDigg(commentIDs []uint) map[uint]int {
	return getBatch(DiggCountCacheKey, commentIDs)
}

func GetAllCacheDigg() map[uint]int {
	return getAll(DiggCountCacheKey)
}

func ClearAllCacheDigg() error {
	return global.Redis.Del(context.Background(), DiggCountCacheKey).Err()
}

func set(key string, commentID uint, increase int) error {
	return global.Redis.HIncrBy(context.Background(), key, strconv.Itoa(int(commentID)), int64(increase)).Err()
}

func get(key string, commentID uint) int {
	num, _ := global.Redis.HGet(context.Background(), key, strconv.Itoa(int(commentID))).Int()
	return num
}

func del(key string, commentID uint) error {
	return global.Redis.HDel(context.Background(), key, strconv.Itoa(int(commentID))).Err()
}

func getBatch(key string, commentIDs []uint) map[uint]int {
	result := make(map[uint]int, len(commentIDs))
	if len(commentIDs) == 0 {
		return result
	}

	fields := make([]string, 0, len(commentIDs))
	for _, commentID := range commentIDs {
		fields = append(fields, strconv.Itoa(int(commentID)))
	}

	values, err := global.Redis.HMGet(context.Background(), key, fields...).Result()
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

func getAll(key string) map[uint]int {
	res, err := global.Redis.HGetAll(context.Background(), key).Result()
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
