// 数据库初始化

package core

import (
	"time"

	"myblogx/global"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/sirupsen/logrus"
)

func InitDB() *gorm.DB {
	// 从配置文件中读取数据库配置
	dbConf := global.Config.DB   // 读库
	dbConf1 := global.Config.DB1 // 写库
	dsn := dbConf.DSN()

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束
	})
	if err != nil {
		logrus.Fatalf("数据库连接失败: %s", err)
	}

	logrus.Infof("数据库连接成功 %s", dsn)

	// 从配置文件中读取 gorm 配置
	gormConf := global.Config.GORM
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(gormConf.MaxIdleConns)
	sqlDB.SetMaxOpenConns(gormConf.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour * time.Duration(gormConf.ConnMaxLifetime))

	if !dbConf1.Empty() {
		// 写库不为空，则注册读写分离的配置
		err := db.Use(dbresolver.Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mysql.Open(dbConf1.DSN())}, // 写
			Replicas: []gorm.Dialector{mysql.Open(dbConf.DSN())},  // 读
			// sources/replicas load balancing policy
			Policy: dbresolver.RandomPolicy{},
		}))
		if err != nil {
			logrus.Fatalf("数据库读写分离配置失败: %s", err)
		}
	}

	return db
}
