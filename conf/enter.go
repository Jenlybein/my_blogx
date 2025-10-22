package conf

type Config struct {
	System System `yaml:"system"`
	Log    Logrus `yaml:"log"`
}
