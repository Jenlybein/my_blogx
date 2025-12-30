package conf

type AI struct {
	Enable    bool   `yaml:"enable" json:"enable"`
	SecretKey string `yaml:"secret" json:"secret"`
	Nickname  string `yaml:"nickname" json:"nickname"`
	Avatar    string `yaml:"avatar" json:"avatar"`
}
