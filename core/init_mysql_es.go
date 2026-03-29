package core

import (
	"myblogx/global"
	"myblogx/service/river_service"
)

func InitMySQLES() {
	if !global.Config.River.Enabled {
		global.Logger.Infof("配置中未启用 MySQL 同步任务")
		return
	}

	r, err := river_service.NewRiver()
	if err != nil {
		global.Logger.Fatal(err)
	}

	go r.Run()
}
