package conf

import "myblogx/service/river_service"

type River struct {
	ServerID uint32        `yaml:"server_id"`
	Flavor   string        `yaml:"flavor"`
	DataDir  string        `yaml:"data_dir"`
	Sources  []RiverSource `yaml:"source"`

	Rules []*river_service.Rule `yaml:"rule"`

	BulkSize int `yaml:"bulk_size"`
}

type RiverSource struct {
}
