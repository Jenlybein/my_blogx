package cron_service

import (
	"context"
	"fmt"
	"myblogx/global"
	"myblogx/models"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const (
	articleSyncLockKey   = "cron:sync_article:lock"
	articleSyncLockTTL   = 30 * time.Minute
	articleSyncMaxRounds = 3
)

var (
	// 原子换桶：将活跃计数桶切到 syncing 桶，避免同步期间新写入被清空或覆盖。
	prepareSyncBucketScript = redis.NewScript(`
if redis.call("EXISTS", KEYS[2]) == 1 then
	return 1
end
if redis.call("EXISTS", KEYS[1]) == 0 then
	return 0
end
redis.call("RENAME", KEYS[1], KEYS[2])
return 1
`)

	// 仅释放自己加的锁，避免误删其他实例的锁。
	releaseLockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)
)

type articleSyncMetric struct {
	name      string
	activeKey string
	syncKey   string
	column    string
}

func SyncArticle() {
	ctx := context.Background()
	unlock, err := lockArticleSync(ctx)
	if err != nil {
		global.Logger.Errorf("同步文章任务获取锁失败 err: %v", err)
		return
	}
	if unlock == nil {
		global.Logger.Infof("同步文章任务跳过，本轮已有任务在执行")
		return
	}
	defer unlock()

	metrics := []articleSyncMetric{
		{name: "收藏数", activeKey: "article_favorite", syncKey: "article_favorite:syncing", column: "favor_count"},
		{name: "点赞数", activeKey: "article_digg", syncKey: "article_digg:syncing", column: "digg_count"},
		{name: "浏览数", activeKey: "article_view", syncKey: "article_view:syncing", column: "view_count"},
	}

	for _, metric := range metrics {
		affected, err := syncArticleMetric(ctx, metric)
		if err != nil {
			global.Logger.Errorf("同步文章任务同步%s失败 key: %s err: %v", metric.name, metric.activeKey, err)
			continue
		}
		if affected > 0 {
			global.Logger.Infof("同步文章任务同步%s成功 key: %s affected: %d", metric.name, metric.activeKey, affected)
		}
	}
}

func lockArticleSync(ctx context.Context) (func(), error) {
	token := strconv.FormatInt(time.Now().UnixNano(), 10)
	locked, err := global.Redis.SetNX(ctx, articleSyncLockKey, token, articleSyncLockTTL).Result()
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, nil
	}

	return func() {
		if _, err := releaseLockScript.Run(ctx, global.Redis, []string{articleSyncLockKey}, token).Result(); err != nil {
			global.Logger.Errorf("同步文章任务释放锁失败 err: %v", err)
		}
	}, nil
}

func syncArticleMetric(ctx context.Context, metric articleSyncMetric) (int, error) {
	totalAffected := 0
	for i := 0; i < articleSyncMaxRounds; i++ {
		hasData, err := prepareSyncBucket(ctx, metric.activeKey, metric.syncKey)
		if err != nil {
			return totalAffected, err
		}
		if !hasData {
			return totalAffected, nil
		}

		affected, err := flushSyncBucket(ctx, metric)
		if err != nil {
			return totalAffected, err
		}
		totalAffected += affected
	}

	global.Logger.Warnf("同步文章任务同步%s达到最大轮次(%d)，剩余增量将在下一轮继续处理", metric.name, articleSyncMaxRounds)
	return totalAffected, nil
}

func prepareSyncBucket(ctx context.Context, activeKey, syncKey string) (bool, error) {
	ret, err := prepareSyncBucketScript.Run(ctx, global.Redis, []string{activeKey, syncKey}).Int()
	if err != nil {
		return false, err
	}
	return ret == 1, nil
}

func flushSyncBucket(ctx context.Context, metric articleSyncMetric) (int, error) {
	rawMap, err := global.Redis.HGetAll(ctx, metric.syncKey).Result()
	if err != nil {
		return 0, err
	}
	if len(rawMap) == 0 {
		if err := global.Redis.Del(ctx, metric.syncKey).Err(); err != nil {
			return 0, err
		}
		return 0, nil
	}

	deltaMap := make(map[uint]int, len(rawMap))
	for articleIDStr, deltaStr := range rawMap {
		articleID, err := strconv.ParseUint(articleIDStr, 10, 64)
		if err != nil {
			global.Logger.Warnf("同步文章任务忽略非法文章ID key: %s article_id: %s", metric.activeKey, articleIDStr)
			continue
		}
		delta, err := strconv.Atoi(deltaStr)
		if err != nil {
			global.Logger.Warnf("同步文章任务忽略非法增量 key: %s article_id: %s delta: %s", metric.activeKey, articleIDStr, deltaStr)
			continue
		}
		if delta == 0 {
			continue
		}
		deltaMap[uint(articleID)] += delta
	}

	if len(deltaMap) == 0 {
		if err := global.Redis.Del(ctx, metric.syncKey).Err(); err != nil {
			return 0, err
		}
		return 0, nil
	}

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		for articleID, delta := range deltaMap {
			if err := applyArticleDelta(tx, metric.column, articleID, delta); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return 0, err
	}

	if err := global.Redis.Del(ctx, metric.syncKey).Err(); err != nil {
		return 0, err
	}

	return len(deltaMap), nil
}

func applyArticleDelta(tx *gorm.DB, column string, articleID uint, delta int) error {
	// 避免并发取消导致点赞/收藏被扣成负数。
	expr := fmt.Sprintf("CASE WHEN %s + ? < 0 THEN 0 ELSE %s + ? END", column, column)
	db := tx.Model(&models.ArticleModel{}).
		Where("id = ?", articleID).
		UpdateColumn(column, gorm.Expr(expr, delta, delta))
	if db.Error != nil {
		return db.Error
	}
	if db.RowsAffected == 0 {
		global.Logger.Warnf("同步文章任务更新行不存在 article_id: %d column: %s delta: %d", articleID, column, delta)
	}
	return nil
}
