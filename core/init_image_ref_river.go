package core

import (
	"myblogx/global"
	"myblogx/service/image_ref_river_service"
)

func InitImageRefRiver() {
	if !global.Config.ImageRefRiver.Enabled {
		global.Logger.Infof("配置中未启用图片引用监听")
		return
	}

	r, err := image_ref_river_service.NewRiver()
	if err != nil {
		global.Logger.Fatal(err)
	}
	go func() {
		if err := r.Run(); err != nil {
			global.Logger.Errorf("图片引用监听运行失败: %v", err)
		}
	}()
}
