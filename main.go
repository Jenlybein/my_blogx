// 主程序入口

package main

import (
	"myblogx/core"
	"myblogx/flags"
	"myblogx/global"
	"myblogx/router"
)

func main() {
	flags.Parse()
	global.Config = core.ReadCfg()
	core.InitLogrus()
	global.DB = core.InitDB()
	global.Redis = core.InitRedis()
	global.ESClient = core.EsConnect()

	flags.Run()

	// 启动 Web 程序
	router.Run()
}
