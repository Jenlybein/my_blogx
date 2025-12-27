// 配置模块入口

package conf

type Config struct {
	System System   `yaml:"system"`
	Jwt    Jwt      `yaml:"jwt"`
	Log    Logrus   `yaml:"log"`
	DB     DB       `yaml:"db"`
	DB1    DB       `yaml:"db1"`
	GORM   GormConf `yaml:"gorm"`
}
