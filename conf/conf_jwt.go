package conf

type Jwt struct {
	Expire int64  `yaml:"expire"`
	Secret string `yaml:"secret"`
	Issuer string `yaml:"issuer"`
}
