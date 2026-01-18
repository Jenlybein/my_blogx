// 主程序入口

package main

import (
	"myblogx/core"
	"myblogx/global"
	"myblogx/router"
)

func main() {
	core.Parse()

	global.Config = core.ReadCfg(&global.Flags.File)
	global.Logger = core.InitLogrus(&global.Config.Log)
	global.DB = core.InitDB(global.Config.DB)
	global.Redis = core.InitRedis(&global.Config.Redis)
	global.ESClient = core.EsConnect(&global.Config.ES)

	core.Run(global.Flags, global.DB)

	// 启动 Web 程序
	router.Run()
}
