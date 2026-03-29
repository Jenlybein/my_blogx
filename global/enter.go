// 全局变量定义

package global

import (
	"database/sql"
	"myblogx/conf"
	"myblogx/service/kafka_service"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v8"
	"github.com/mojocn/base64Captcha"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const Version = "1.0.0"

type FlagRecord struct {
	File string
}

var (
	Flags             *FlagRecord
	Config            *conf.Config
	Logger            *logrus.Logger
	DB                *gorm.DB
	ClickHouse        *sql.DB
	Redis             *redis.Client
	KafkaMysqlClient  *kafka_service.KafkaMysqlClient
	ESClient          *elasticsearch.Client
	ImageCaptchaStore = base64Captcha.DefaultMemStore
)
