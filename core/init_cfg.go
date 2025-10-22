package core

import (
	"fmt"
	"os"

	"myblogx/conf"
	"myblogx/flags"

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

	fmt.Println("读取配置文件 %s 成功", settings)

	return c
}
