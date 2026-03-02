package cron_service

import (
	"context"
	"fmt"
	"myblogx/global"
	"myblogx/models"
	"myblogx/service/redis_service"
	"myblogx/service/redis_service/redis_article"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const (
	// 定时任务分布式锁 key，防止多实例并发执行同一个同步任务。
	articleSyncLockKey = "cron:sync_article:lock"
	// 锁超时时间，避免进程异常退出后死锁。
	articleSyncLockTTL = 30 * time.Minute
)

var (
	// 原子换桶脚本：
	// 1) 如果 syncing 桶已存在，说明有上一批待处理数据，直接返回可处理；
	// 2) 如果 active 桶不存在，说明没有增量；
	// 3) 否则把 active 原子 rename 到 syncing，切出一个稳定快照用于刷库。
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
)

// articleSyncMetric 描述一类需要同步的统计指标。
type articleSyncMetric struct {
	// name 用于日志显示。
	name string
	// activeKey 是在线写入的 Redis Hash key。
	activeKey string
	// syncKey 是被切出的待同步 Redis Hash key。
	syncKey string
	// column 是数据库文章表对应的计数字段。
	column string
}

// SyncArticle 定时任务入口：把 Redis 增量同步回数据库。
func SyncArticle() {
	// 使用统一 context。
	ctx := context.Background()

	// 先抢锁，避免多实例重复同步。
	unlock, err := redis_service.LockArticleSync(ctx, articleSyncLockKey, articleSyncLockTTL)
	if err != nil {
		global.Logger.Errorf("同步文章任务获取锁失败 err: %v", err)
		return
	}

	// 没拿到锁则说明已有实例在执行，本轮直接跳过。
	if unlock == nil {
		global.Logger.Infof("同步文章任务跳过，本轮已有任务在执行")
		return
	}

	// 任务结束后释放锁。
	defer unlock()

	// 定义统计指标与 Redis/DB 的映射关系。
	metrics := []articleSyncMetric{
		{name: "收藏数", activeKey: string(redis_article.ArticleCacheFavorite), syncKey: "article_favorite:syncing", column: "favor_count"},
		{name: "点赞数", activeKey: string(redis_article.ArticleCacheDigg), syncKey: "article_digg:syncing", column: "digg_count"},
		{name: "浏览数", activeKey: string(redis_article.ArticleCacheView), syncKey: "article_view:syncing", column: "view_count"},
		{name: "评论数", activeKey: string(redis_article.ArticleCacheComment), syncKey: "article_comment:syncing", column: "comment_count"},
	}

	// 逐类同步，互不影响；某一类失败不阻塞其他类。
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

// syncArticleMetric 同步某一个指标（收藏/点赞/浏览）。
func syncArticleMetric(ctx context.Context, metric articleSyncMetric) (int, error) {
	// 原子切桶：把 active 变成 syncing，得到本轮稳定数据。(执行桶改名脚本，返回是否有数据可同步。)
	ret, err := prepareSyncBucketScript.Run(ctx, global.Redis, []string{metric.activeKey, metric.syncKey}).Int()
	if err != nil {
		return 0, err
	}
	// 没数据可处理则提前结束。
	if ret != 1 {
		return 0, nil
	}

	// 将 syncing 桶的数据同步到数据库。
	affected, err := flushSyncBucket(ctx, metric)
	if err != nil {
		return 0, err
	}

	// 返回累计本轮影响文章数。
	return affected, nil
}

// flushSyncBucket 把一个 syncing 桶里的增量落库，并在成功后删除该桶。
func flushSyncBucket(ctx context.Context, metric articleSyncMetric) (int, error) {
	// 读取 syncing 桶全部字段（article_id -> delta）。
	rawMap, err := global.Redis.HGetAll(ctx, metric.syncKey).Result()
	if err != nil {
		return 0, err
	}

	// 空桶直接删除，避免脏 key 残留。
	if len(rawMap) == 0 {
		if err := global.Redis.Del(ctx, metric.syncKey).Err(); err != nil {
			return 0, err
		}
		return 0, nil
	}

	// 解析并规整增量，键是文章 ID，值是净增量。
	deltaMap := make(map[uint]int, len(rawMap))
	for articleIDStr, deltaStr := range rawMap {
		// 解析文章 ID。
		articleID, err := strconv.ParseUint(articleIDStr, 10, 64)
		if err != nil {
			global.Logger.Warnf("同步文章任务忽略非法文章ID key: %s article_id: %s", metric.activeKey, articleIDStr)
			continue
		}

		// 解析增量值。
		delta, err := strconv.Atoi(deltaStr)
		if err != nil {
			global.Logger.Warnf("同步文章任务忽略非法增量 key: %s article_id: %s delta: %s", metric.activeKey, articleIDStr, deltaStr)
			continue
		}

		// 0 增量没意义，直接跳过。
		if delta == 0 {
			continue
		}

		// 合并同一文章的增量。
		deltaMap[uint(articleID)] += delta
	}

	// 解析后无有效数据则删桶结束。
	if len(deltaMap) == 0 {
		if err := global.Redis.Del(ctx, metric.syncKey).Err(); err != nil {
			return 0, err
		}
		return 0, nil
	}

	// 循环写库
	for articleID, delta := range deltaMap {
		if err := applyArticleDelta(metric.column, articleID, delta); err != nil {
			global.Logger.Warnf("同步文章任务同步%s失败，准备回补active key: %s article_id: %d delta: %d err: %v", metric.name, metric.activeKey, articleID, delta, err)
			if requeueErr := global.Redis.HIncrBy(ctx, metric.activeKey, strconv.Itoa(int(articleID)), int64(delta)).Err(); requeueErr != nil {
				global.Logger.Errorf("同步文章任务回补active失败 key: %s article_id: %d delta: %d err: %v", metric.activeKey, articleID, delta, requeueErr)
			}
		}
	}

	// 写库成功后删除 syncing 桶，表示这批数据已消费完成。
	if err := global.Redis.Del(ctx, metric.syncKey).Err(); err != nil {
		return 0, err
	}

	// 返回本次实际更新的文章数量。
	return len(deltaMap), nil
}

// applyArticleDelta 对单篇文章执行增量更新。
func applyArticleDelta(column string, articleID uint, delta int) error {
	// 使用 CASE 防止减到负数（如点赞/收藏取消过多）。
	expr := fmt.Sprintf("CASE WHEN %s + ? < 0 THEN 0 ELSE %s + ? END", column, column)

	// UpdateColumn 使用数据库表达式原子更新，避免先读后写竞争。
	db := global.DB.Model(&models.ArticleModel{}).
		Where("id = ?", articleID).
		UpdateColumn(column, gorm.Expr(expr, delta, delta))

	// SQL 执行失败直接返回。
	if db.Error != nil {
		return db.Error
	}

	// 如果文章不存在，记录告警但不作为错误中断整批任务。
	if db.RowsAffected == 0 {
		global.Logger.Warnf("同步文章任务更新行不存在 article_id: %d column: %s delta: %d", articleID, column, delta)
	}
	return nil
}
