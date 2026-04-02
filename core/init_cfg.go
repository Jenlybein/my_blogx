// 配置初始化

package core

import (
	"fmt"
	"os"

	"myblogx/conf"
	"myblogx/global"
	"myblogx/utils/envyaml"

	"gopkg.in/yaml.v3"
)

func ReadCfg(settings *string) (c *conf.Config) {
	byteData, err := os.ReadFile(*settings)
	if err != nil {
		panic(err)
	}

	c = new(conf.Config)

	err = envyaml.Unmarshal(byteData, c)

	if err != nil {
		panic(fmt.Errorf("yaml 配置文件解析失败: %s", err))
	}

	fmt.Printf("读取配置文件 %s 成功\n", *settings)

	return c
}

func SetCfg(cfg *conf.Config, settings *string) {
	byteData, err := yaml.Marshal(*cfg)
	if err != nil {
		global.Logger.Errorf("yaml 配置文件序列化失败: %s", err)
	}

	err = os.WriteFile(*settings, byteData, 0666)
	if err != nil {
		global.Logger.Errorf("yaml 配置文件写入失败: %s", err)
	}
}
