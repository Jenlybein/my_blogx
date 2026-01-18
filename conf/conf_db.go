// 数据库配置

package conf

import (
	"fmt"
)

type DB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Debug    bool   `yaml:"debug"`  // 是否开启调试模式：开启后会打印 SQL 语句
	Source   string `yaml:"source"` // 数据库的源
}

func (d *DB) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName)
}

func (d DB) Empty() bool {
	return d.User == "" && d.Password == "" && d.Host == "" && d.Port == 0 && d.DBName == ""
}
