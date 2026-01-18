// 数据库初始化

package core

import (
	"time"

	"myblogx/conf"
	"myblogx/global"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func InitDB(dbCfg []conf.DB) *gorm.DB {
	if len(dbCfg) == 0 {
		global.Logger.Fatalf("数据库配置错误：未配置数据库")
	}

	// 从配置文件中读取数据库配置
	DB := dbCfg[0] // 写库
	dsn := DB.DSN()

	// 连接数据库（使用主库初始化）
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束
	})
	if err != nil {
		global.Logger.Fatalf("数据库连接失败: %s", err)
	}

	global.Logger.Infof("数据库连接成功 %s", dsn)

	// 从配置文件中读取 gorm 配置
	gormConf := global.Config.GORM
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(gormConf.MaxIdleConns)
	sqlDB.SetMaxOpenConns(gormConf.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour * time.Duration(gormConf.ConnMaxLifetime))

	if len(dbCfg) > 1 {
		var readList []gorm.Dialector
		for _, d := range dbCfg[1:] {
			readList = append(readList, mysql.Open(d.DSN()))
		}
		// 读库不为空，则注册读写分离的配置
		err := db.Use(dbresolver.Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mysql.Open(DB.DSN())}, // 写库（主库）
			Replicas: readList,                               // 读库（从库）
			Policy:   dbresolver.RandomPolicy{},
		}))
		if err != nil {
			global.Logger.Fatalf("数据库读写分离配置失败: %s", err)
		}
		global.Logger.Infof("数据库读写分离配置成功 %d 个读库", len(readList))
	}

	return db
}
