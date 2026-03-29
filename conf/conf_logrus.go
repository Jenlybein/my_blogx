// 日志配置

package conf

type Logrus struct {
	App               string `yaml:"app"`
	Dir               string `yaml:"dir"`
	Level             string `yaml:"level"`
	StdoutFormat      string `yaml:"stdout_format"`       // text/json，默认 json
	RequestLogEnabled bool   `yaml:"request_log_enabled"` // 是否记录请求访问日志
	QueryDefaultLimit int    `yaml:"query_default_limit"` // 后台日志查询默认条数
	QueryMaxLimit     int    `yaml:"query_max_limit"`     // 后台日志查询最大条数
}
