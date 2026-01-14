// 配置模块入口

package conf

type Config struct {
	System System   `yaml:"system"`
	Jwt    Jwt      `yaml:"jwt"`
	Log    Logrus   `yaml:"log"`
	DB     DBConfig `yaml:"db"`
	GORM   GormConf `yaml:"gorm"`
	Redis  Redis    `yaml:"redis"`
	Upload Upload   `yaml:"upload"`
	Site   Site     `yaml:"site"`
	Email  Email    `yaml:"email"`
	QQ     QQ       `yaml:"qq"`
	QiNiu  QiNiu    `yaml:"qiniu"`
	AI     AI       `yaml:"ai"`
}
