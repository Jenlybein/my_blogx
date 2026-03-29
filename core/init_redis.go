package core

import (
	"context"
	"myblogx/conf"
	"myblogx/global"

	"github.com/go-redis/redis/v8"
)

func InitRedis(redisCfg *conf.Redis) *redis.Client {
	redisDB := redis.NewClient(&redis.Options{
		Addr:     redisCfg.GetAddr(),
		Username: redisCfg.Username,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})

	_, err := redisDB.Ping(context.Background()).Result()
	if err != nil {
		global.Logger.Fatalf("Redis 连接失败: %v", err)
	}

	global.Logger.Infof("Redis 连接成功: %s", redisDB.Options().Addr)

	return redisDB
}
