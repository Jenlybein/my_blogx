package core

import (
	"myblogx/global"
	"myblogx/service/river_service"
)

func InitMySQLES() {
	if !global.Config.River.Enabled {
		global.Logger.Infof("配置中未启用mysql同步操作")
		return
	}

	r, err := river_service.NewRiver()
	if err != nil {
		global.Logger.Fatal(err)
	}

	go r.Run()
}
