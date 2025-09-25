package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var cfgPath = "settings.yaml"

type System struct {
	IP   string `yaml:"ip"`
	Port int    `yaml:"port"`
}
type Config struct {
	System System `yaml:"system"`
}

func ReadCfg() {
	byteData, err := os.ReadFile(cfgPath)
	if err != nil {
		panic(err)
	}

	var config Config

	err = yaml.Unmarshal(byteData, &config)

	if err != nil {
		panic(fmt.Errorf("yaml 配置文件解析失败: %s", err))
	}

	fmt.Printf("%+v\n", config)
}