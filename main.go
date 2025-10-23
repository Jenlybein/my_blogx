package main

import (
	"myblogx/core"
	"myblogx/flags"
	"myblogx/global"
)

func main() {
	flags.Parse()
	global.Config = core.ReadCfg()
	core.InitLogrus()
	global.DB = core.InitDB()
}
