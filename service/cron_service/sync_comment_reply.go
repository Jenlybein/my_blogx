package cron_service

import (
	"context"
	"fmt"
	"myblogx/global"
	"myblogx/models"
	"myblogx/service/redis_service"
	"myblogx/service/redis_service/redis_comment"
	"strconv"
	"time"

	"gorm.io/gorm"
)

const (
	commentReplySyncLockKey = "cron:sync_comment_reply:lock"
	commentReplySyncLockTTL = 30 * time.Minute
	commentReplySyncingKey  = "comment_reply:syncing"
)

func SyncCommentReply() {
	ctx := context.Background()

	unlock, err := redis_service.LockArticleSync(ctx, commentReplySyncLockKey, commentReplySyncLockTTL)
	if err != nil {
		global.Logger.Errorf("同步评论回复数任务获取锁失败 err: %v", err)
		return
	}
	if unlock == nil {
		global.Logger.Infof("同步评论回复数任务跳过，本轮已有任务在执行")
		return
	}
	defer unlock()

	affected, err := syncCommentReplyMetric(ctx)
	if err != nil {
		global.Logger.Errorf("同步评论回复数任务失败 err: %v", err)
		return
	}
	if affected > 0 {
		global.Logger.Infof("同步评论回复数任务成功 affected: %d", affected)
	}
}

func syncCommentReplyMetric(ctx context.Context) (int, error) {
	ret, err := prepareSyncBucketScript.Run(ctx, global.Redis, []string{redis_comment.ReplyCountCacheKey, commentReplySyncingKey}).Int()
	if err != nil {
		return 0, err
	}
	if ret != 1 {
		return 0, nil
	}

	rawMap, err := global.Redis.HGetAll(ctx, commentReplySyncingKey).Result()
	if err != nil {
		return 0, err
	}
	if len(rawMap) == 0 {
		if err := global.Redis.Del(ctx, commentReplySyncingKey).Err(); err != nil {
			return 0, err
		}
		return 0, nil
	}

	deltaMap := make(map[uint]int, len(rawMap))
	for commentIDStr, deltaStr := range rawMap {
		commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
		if err != nil {
			global.Logger.Warnf("同步评论回复数任务忽略非法评论ID comment_id: %s", commentIDStr)
			continue
		}
		delta, err := strconv.Atoi(deltaStr)
		if err != nil {
			global.Logger.Warnf("同步评论回复数任务忽略非法增量 comment_id: %s delta: %s", commentIDStr, deltaStr)
			continue
		}
		if delta == 0 {
			continue
		}
		deltaMap[uint(commentID)] += delta
	}

	if len(deltaMap) == 0 {
		if err := global.Redis.Del(ctx, commentReplySyncingKey).Err(); err != nil {
			return 0, err
		}
		return 0, nil
	}

	for commentID, delta := range deltaMap {
		if err := applyCommentReplyDelta(commentID, delta); err != nil {
			global.Logger.Warnf("同步评论回复数任务写库失败，准备回补缓存 comment_id: %d delta: %d err: %v", commentID, delta, err)
			if requeueErr := global.Redis.HIncrBy(ctx, redis_comment.ReplyCountCacheKey, strconv.Itoa(int(commentID)), int64(delta)).Err(); requeueErr != nil {
				global.Logger.Errorf("同步评论回复数任务回补缓存失败 comment_id: %d delta: %d err: %v", commentID, delta, requeueErr)
			}
		}
	}

	if err := global.Redis.Del(ctx, commentReplySyncingKey).Err(); err != nil {
		return 0, err
	}
	return len(deltaMap), nil
}

func applyCommentReplyDelta(commentID uint, delta int) error {
	expr := fmt.Sprintf("CASE WHEN %s + ? < 0 THEN 0 ELSE %s + ? END", "reply_count", "reply_count")

	db := global.DB.Model(&models.CommentModel{}).
		Where("id = ?", commentID).
		UpdateColumn("reply_count", gorm.Expr(expr, delta, delta))
	if db.Error != nil {
		return db.Error
	}
	if db.RowsAffected == 0 {
		global.Logger.Warnf("同步评论回复数任务更新行不存在 comment_id: %d delta: %d", commentID, delta)
	}
	return nil
}
