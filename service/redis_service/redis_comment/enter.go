package redis_comment

import (
	"context"
	"fmt"
	"myblogx/global"
	"myblogx/models/ctype"
	"strconv"
)

const ReplyCountCacheKey = "comment_reply"
const DiggCountCacheKey = "comment_digg"

func SetCacheReply(commentID ctype.ID, increase int) error {
	return set(ReplyCountCacheKey, commentID, increase)
}

func GetCacheReply(commentID ctype.ID) int {
	return get(ReplyCountCacheKey, commentID)
}

func DelCacheReply(commentID ctype.ID) error {
	return del(ReplyCountCacheKey, commentID)
}

func GetBatchCacheReply(commentIDs []ctype.ID) map[ctype.ID]int {
	return getBatch(ReplyCountCacheKey, commentIDs)
}

func GetAllCacheReply() map[ctype.ID]int {
	return getAll(ReplyCountCacheKey)
}

func ClearAllCacheReply() error {
	return global.Redis.Del(context.Background(), ReplyCountCacheKey).Err()
}

func SetCacheDigg(commentID ctype.ID, increase int) error {
	return set(DiggCountCacheKey, commentID, increase)
}

func GetCacheDigg(commentID ctype.ID) int {
	return get(DiggCountCacheKey, commentID)
}

func DelCacheDigg(commentID ctype.ID) error {
	return del(DiggCountCacheKey, commentID)
}

func GetBatchCacheDigg(commentIDs []ctype.ID) map[ctype.ID]int {
	return getBatch(DiggCountCacheKey, commentIDs)
}

func GetAllCacheDigg() map[ctype.ID]int {
	return getAll(DiggCountCacheKey)
}

func ClearAllCacheDigg() error {
	return global.Redis.Del(context.Background(), DiggCountCacheKey).Err()
}

func set(key string, commentID ctype.ID, increase int) error {
	return global.Redis.HIncrBy(context.Background(), key, commentID.String(), int64(increase)).Err()
}

func get(key string, commentID ctype.ID) int {
	num, _ := global.Redis.HGet(context.Background(), key, commentID.String()).Int()
	return num
}

func del(key string, commentID ctype.ID) error {
	return global.Redis.HDel(context.Background(), key, commentID.String()).Err()
}

func getBatch(key string, commentIDs []ctype.ID) map[ctype.ID]int {
	result := make(map[ctype.ID]int, len(commentIDs))
	if len(commentIDs) == 0 {
		return result
	}

	fields := make([]string, 0, len(commentIDs))
	for _, commentID := range commentIDs {
		fields = append(fields, commentID.String())
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

func getAll(key string) map[ctype.ID]int {
	res, err := global.Redis.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil
	}

	numMap := make(map[ctype.ID]int, len(res))
	for k, v := range res {
		commentID, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			continue
		}
		num, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		numMap[ctype.ID(commentID)] = num
	}
	return numMap
}
