// 主程序入口

package main

import (
	"myblogx/core"
	"myblogx/flags"
	"myblogx/global"
	"myblogx/router"
)

func main() {
	flag := flags.Parse()

	global.Flags = &global.FlagRecord{
		File: flag.File,
	}

	global.Config = core.ReadCfg(&flag.File)
	global.Logger = core.InitLogrus(&global.Config.Log)
	global.Redis = core.InitRedis(&global.Config.Redis)
	global.DB = core.InitDB(global.Config.DB)
	global.ESClient = core.EsConnect(&global.Config.ES)

	flags.Run(flag, global.DB)

	core.InitMySQLES()

	// 启动 Web 程序
	router.Run()
}
