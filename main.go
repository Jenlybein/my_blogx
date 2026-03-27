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
	global.Logger = core.InitLogrus(&global.Config.Log)
	if err := core.InitSnowflake(); err != nil {
		panic(err)
	}
	global.Redis = core.InitRedis(&global.Config.Redis)
	// global.KafkaMysqlClient = core.KafkaMysqlClientInit(&global.Config.Kafka)
	global.DB = core.InitDB(global.Config.DB)
	global.ESClient = core.EsConnect(&global.Config.ES)

	flags.Run(flag, global.DB)

	core.InitMySQLES()
	core.InitImageRefRiver()

	// 定时任务
	cron_service.Cron()

	// 启动 Web 程序
	router.Run()
}
