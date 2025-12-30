// 全局变量定义

package global

import (
	"myblogx/conf"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const Version = "1.0.0"

var (
	Config *conf.Config
	DB     *gorm.DB
	Redis  *redis.Client
)
