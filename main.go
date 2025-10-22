package main

import (
	"myblogx/core"
	"myblogx/flags"
	"myblogx/global"

	"github.com/sirupsen/logrus"
)

func main() {
	flags.Parse()
	global.Config = core.ReadCfg()
	core.InitLogrus()

	logrus.Debug("123")
	logrus.Infof("日志初始化完成")
	logrus.Error("456")
	logrus.Warnf("789")
}
