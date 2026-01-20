package river_service

// import (
// 	"io/ioutil"
// 	"time"

// 	"github.com/BurntSushi/toml"
// 	"github.com/pingcap/errors"
// )

// // SourceConfig 定义数据源的配置信息
// type SourceConfig struct {
// 	Schema string   `toml:"schema"` // 数据库模式名称
// 	Tables []string `toml:"tables"` // 需要同步的表列表
// }

// // Config 定义River服务的主要配置参数
// type Config struct {
// 	// MySQL连接配置
// 	MyAddr     string `toml:"my_addr"`    // MySQL服务器地址
// 	MyUser     string `toml:"my_user"`    // MySQL用户名
// 	MyPassword string `toml:"my_pass"`    // MySQL密码
// 	MyCharset  string `toml:"my_charset"` // MySQL字符集

// 	// Elasticsearch连接配置
// 	ESHttps    bool   `toml:"es_https"`   // 是否使用HTTPS连接Elasticsearch
// 	ESAddr     string `toml:"es_addr"`    // Elasticsearch服务器地址
// 	ESUser     string `toml:"es_user"`    // Elasticsearch用户名
// 	ESPassword string `toml:"es_pass"`    // Elasticsearch密码

// 	// 统计信息配置
// 	StatAddr string `toml:"stat_addr"`    // 统计信息服务器地址
// 	StatPath string `toml:"stat_path"`    // 统计信息路径

// 	// MySQL复制配置
// 	ServerID uint32 `toml:"server_id"`    // MySQL复制的Server ID
// 	Flavor   string `toml:"flavor"`       // MySQL版本类型
// 	DataDir  string `toml:"data_dir"`     // 数据存储目录

// 	// 导出配置
// 	DumpExec       string `toml:"mysqldump"`      // mysqldump命令路径
// 	SkipMasterData bool   `toml:"skip_master_data"` // 是否跳过主数据

// 	// 数据源配置
// 	Sources []SourceConfig `toml:"source"` // 数据源列表

// 	// 规则配置
// 	Rules []*Rule `toml:"rule"` // 同步规则列表

// 	// 批量处理配置
// 	BulkSize int `toml:"bulk_size"` // 批量处理大小

// 	// 刷新时间配置
// 	FlushBulkTime TomlDuration `toml:"flush_bulk_time"` // 批量刷新时间间隔

// 	// 表处理配置
// 	SkipNoPkTable bool `toml:"skip_no_pk_table"` // 是否跳过没有主键的表
// }

// // NewConfigWithFile 从文件创建一个新的配置对象
// func NewConfigWithFile(name string) (*Config, error) {
// 	data, err := ioutil.ReadFile(name)
// 	if err != nil {
// 		return nil, errors.Trace(err)
// 	}

// 	return NewConfig(string(data))
// }

// // NewConfig 从数据字符串创建一个新的配置对象
// func NewConfig(data string) (*Config, error) {
// 	var c Config

// 	_, err := toml.Decode(data, &c)
// 	if err != nil {
// 		return nil, errors.Trace(err)
// 	}

// 	return &c, nil
// }

// // TomlDuration 支持TOML格式的时间编解码
// type TomlDuration struct {
// 	time.Duration
// }

// // UnmarshalText 实现TOML的UnmarshalText接口
// func (d *TomlDuration) UnmarshalText(text []byte) error {
// 	var err error
// 	d.Duration, err = time.ParseDuration(string(text))
// 	return err
// }
//
