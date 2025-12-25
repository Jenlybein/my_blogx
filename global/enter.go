// 全局变量定义

package global

import (
	"myblogx/conf"

	"gorm.io/gorm"
)

var (
	Config *conf.Config
	DB     *gorm.DB
)
