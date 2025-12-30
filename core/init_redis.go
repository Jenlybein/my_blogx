package core

import (
	"context"
	"myblogx/global"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

func InitRedis() *redis.Client {
	redisDB := redis.NewClient(&redis.Options{
		Addr:     global.Config.Redis.GetAddr(),
		Username: global.Config.Redis.Username,
		Password: global.Config.Redis.Password,
		DB:       global.Config.Redis.DB,
	})

	_, err := redisDB.Ping(context.Background()).Result()
	if err != nil {
		logrus.Fatalf("redis 连接失败: %v", err)
	}

	logrus.Infof("redis 连接成功: %s", redisDB.Options().Addr)

	return redisDB
}
