// flags/enter.go
package core

import (
	"flag"
	"fmt"
	"myblogx/global"
	"myblogx/service/flag_service"
	"os"

	"gorm.io/gorm"
)

func Parse() {
	var Flags = new(global.FlagOptions)

	flag.StringVar(&Flags.File, "f", "settings.yaml", "指定配置文件路径")
	flag.BoolVar(&Flags.DB, "db", false, "数据库迁移")
	flag.BoolVar(&Flags.Version, "version", false, "显示版本信息")
	flag.StringVar(&Flags.Type, "t", "", "操作类型")
	flag.StringVar(&Flags.Sub, "s", "", "子操作类型")

	flag.Parse()

	global.Flags = Flags
}

func Run(op *global.FlagOptions, db *gorm.DB) {
	if op.DB {
		// 执行数据库迁移
		flag_service.FlagDB(db)
		os.Exit(0)
	}

	switch op.Type {
	case "user":
		u := flag_service.FlagUser{}
		switch op.Sub {
		case "create":
			u.Create(db)
			os.Exit(0)
		default:
			fmt.Println("未知子操作类型")
			os.Exit(1)
		}
	}
}
