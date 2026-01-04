// 配置初始化

package core

import (
	"fmt"
	"os"

	"myblogx/conf"
	"myblogx/flags"
	"myblogx/global"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func ReadCfg() (c *conf.Config) {
	settings := flags.FlagOptions.File
	byteData, err := os.ReadFile(settings)
	if err != nil {
		panic(err)
	}

	c = new(conf.Config)

	err = yaml.Unmarshal(byteData, &c)

	if err != nil {
		panic(fmt.Errorf("yaml 配置文件解析失败: %s", err))
	}

	fmt.Printf("读取配置文件 %s 成功\n", settings)

	return c
}

func SetCfg() {
	byteData, err := yaml.Marshal(global.Config)
	if err != nil {
		logrus.Errorf("yaml 配置文件序列化失败: %s", err)
	}

	err = os.WriteFile(flags.FlagOptions.File, byteData, 0666)
	if err != nil {
		logrus.Errorf("yaml 配置文件写入失败: %s", err)
	}

	global.Config = ReadCfg()
}
