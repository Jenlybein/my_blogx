package core

import (
	"fmt"
	"os"

	"myblogx/flags"

	"gopkg.in/yaml.v3"
)

type System struct {
	IP   string `yaml:"ip"`
	Port int    `yaml:"port"`
}
type Config struct {
	System System `yaml:"system"`
}

func ReadCfg() {
	settings := flags.FlagOptions.File
	byteData, err := os.ReadFile(settings)
	if err != nil {
		panic(err)
	}

	var config Config

	err = yaml.Unmarshal(byteData, &config)

	if err != nil {
		panic(fmt.Errorf("yaml 配置文件解析失败: %s", err))
	}

	fmt.Printf("读取配置文件 %s 成功", settings)
}
