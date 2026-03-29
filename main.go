// 主程序入口

package main

import (
	"myblogx/core"
	"myblogx/flags"
	"myblogx/global"
	"myblogx/router"
	"myblogx/service/cron_service"
)

func main() {
	flag := flags.Parse()

	global.Flags = &global.FlagRecord{
		File: flag.File,
	}

	global.Config = core.ReadCfg(&flag.File)
	if err := core.InitSnowflake(); err != nil {
		panic(err)
	}
	global.Logger = core.InitLogrus(&global.Config.Log)
	global.Redis = core.InitRedis(&global.Config.Redis)
	// global.KafkaMysqlClient = core.KafkaMysqlClientInit(&global.Config.Kafka)
	global.DB = core.InitDB(global.Config.DB)
	global.ClickHouse = core.InitClickHouse(&global.Config.ClickHouse)
	global.ESClient = core.EsConnect(&global.Config.ES)

	flags.Run(flag, global.DB)
	if err := core.InitRuntimeSite(); err != nil {
		panic(err)
	}

	core.InitMySQLES()
	core.InitImageRefRiver()

	// 定时任务
	cron_service.Cron()

	// 启动 Web 程序
	router.Run()
}
