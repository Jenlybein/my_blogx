// Elasticsearch配置

package conf

type ES struct {
	Addresses []string `yaml:"addresses"` // ES 地址列表
	Username  string   `yaml:"username"`  // ES 用户名
	Password  string   `yaml:"password"`  // ES 密码
	Index     string   `yaml:"index"`     // 默认索引名
}
