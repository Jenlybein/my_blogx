package cron_service

import (
	"context"
	"strconv"
	"time"

	"myblogx/global"
	"myblogx/models/ctype"
	"myblogx/service/redis_service"

	"github.com/go-co-op/gocron/v2"
	"github.com/go-redis/redis/v8"
)

type CronService struct{}

var (
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

type hashCounterSyncConfig struct {
	taskName   string
	metricName string
	activeKey  string
	syncKey    string
	idName     string
	applyDelta func(id ctype.ID, delta int) error
}

func syncCounters() {
	SyncArticle()
	SyncTag()
	SyncCommentReply()
	SyncCommentDigg()
}

func runLockedSyncTask(taskName, lockKey string, lockTTL time.Duration, syncFunc func(ctx context.Context) (int, error)) {
	ctx := context.Background()

	unlock, err := redis_service.LockArticleSync(ctx, lockKey, lockTTL)
	if err != nil {
		global.Logger.Errorf("%s获取锁失败: 错误=%v", taskName, err)
		return
	}
	if unlock == nil {
		global.Logger.Infof("%s跳过，本轮已有任务在执行", taskName)
		return
	}
	defer unlock()

	affected, err := syncFunc(ctx)
	if err != nil {
		global.Logger.Errorf("%s执行失败: 错误=%v", taskName, err)
		return
	}
	if affected > 0 {
		global.Logger.Infof("%s执行成功: 影响数量=%d", taskName, affected)
	}
}

func syncHashCounterMetric(ctx context.Context, config hashCounterSyncConfig) (int, error) {
	logPrefix := config.taskName
	if config.metricName != "" {
		logPrefix = config.taskName + "同步" + config.metricName
	}

	ret, err := prepareSyncBucketScript.Run(ctx, global.Redis, []string{config.activeKey, config.syncKey}).Int()
	if err != nil {
		return 0, err
	}
	if ret != 1 {
		return 0, nil
	}

	rawMap, err := global.Redis.HGetAll(ctx, config.syncKey).Result()
	if err != nil {
		return 0, err
	}
	if len(rawMap) == 0 {
		if err := global.Redis.Del(ctx, config.syncKey).Err(); err != nil {
			return 0, err
		}
		return 0, nil
	}

	deltaMap := make(map[ctype.ID]int, len(rawMap))
	for idStr, deltaStr := range rawMap {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			global.Logger.Warnf("%s忽略非法%s: Redis键=%s %s=%s", logPrefix, config.idName, config.activeKey, config.idName, idStr)
			continue
		}

		delta, err := strconv.Atoi(deltaStr)
		if err != nil {
			global.Logger.Warnf("%s忽略非法增量: Redis键=%s %s=%s 增量=%s", logPrefix, config.activeKey, config.idName, idStr, deltaStr)
			continue
		}

		if delta == 0 {
			continue
		}
		deltaMap[ctype.ID(id)] += delta
	}

	if len(deltaMap) == 0 {
		if err := global.Redis.Del(ctx, config.syncKey).Err(); err != nil {
			return 0, err
		}
		return 0, nil
	}

	for id, delta := range deltaMap {
		if err := config.applyDelta(id, delta); err != nil {
			global.Logger.Warnf("%s写库失败，准备回补缓存: Redis键=%s %s=%d 增量=%d 错误=%v", logPrefix, config.activeKey, config.idName, id, delta, err)
			if requeueErr := global.Redis.HIncrBy(ctx, config.activeKey, strconv.FormatUint(uint64(id), 10), int64(delta)).Err(); requeueErr != nil {
				global.Logger.Errorf("%s回补缓存失败: Redis键=%s %s=%d 增量=%d 错误=%v", logPrefix, config.activeKey, config.idName, id, delta, requeueErr)
			}
		}
	}

	if err := global.Redis.Del(ctx, config.syncKey).Err(); err != nil {
		return 0, err
	}
	return len(deltaMap), nil
}

func Cron() {
	timezone, _ := time.LoadLocation("Asia/Shanghai")
	s, err := gocron.NewScheduler(
		gocron.WithLocation(timezone),
	)
	if err != nil {
		global.Logger.Errorf("创建 gocron 调度器失败: %v", err)
	}

	_, err = s.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(2, 0, 0))),
		gocron.NewTask(syncCounters),
	)
	if err != nil {
		global.Logger.Errorf("添加同步任务失败: %v", err)
	}

	global.Logger.Infof("成功启动定时任务")
	s.Start()
}
