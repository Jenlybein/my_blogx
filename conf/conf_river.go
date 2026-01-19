package conf

import "myblogx/service/river_service/rule"

type River struct {
	Enabled bool `yaml:"enabled"`

	ServerID uint32        `yaml:"server_id"`
	Flavor   string        `yaml:"flavor"`
	DataDir  string        `yaml:"data_dir"`
	Sources  []RiverSource `yaml:"source"`

	Rules []*rule.Rule `yaml:"rule"`

	Charset string `yaml:"charset"`

	DumpExec       string `yaml:"mysqldump"`
	SkipMasterData bool   `yaml:"skip_master_data"`

	BulkSize      int `yaml:"bulk_size"`
	FlushBulkTime int `yaml:"flush_bulk_time"`
}

type RiverSource struct {
	Schema string   `yaml:"schema"`
	Tables []string `yaml:"tables"`
}
