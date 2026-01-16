// 全局变量定义

package global

import (
	"myblogx/conf"
	"myblogx/store/email_store"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v8"
	"github.com/mojocn/base64Captcha"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const Version = "1.0.0"

type FlagOptions struct {
	/* 定义命令行参数选项结构体,用于存储和处理命令行传入的各种标志参数 */
	File    string
	DB      bool
	Version bool
	Type    string
	Sub     string
}

var (
	Flags             *FlagOptions
	Config            *conf.Config
	Logger            *logrus.Logger
	DB                *gorm.DB
	Redis             *redis.Client
	ESClient          *elasticsearch.Client
	ImageCaptchaStore = base64Captcha.DefaultMemStore
	EmailVerifyStore  = email_store.NewEmailVerifyStore(3, 5)
)
