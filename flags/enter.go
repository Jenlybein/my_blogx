// flags/enter.go
package flags

import (
	"flag"
	"fmt"
	"os"

	"gorm.io/gorm"
)

type FlagOptions struct {
	/* 定义命令行参数选项结构体,用于存储和处理命令行传入的各种标志参数 */
	File    string
	DB      bool
	Version bool
	Type    string
	Sub     string
	ES      bool
}

func Parse() *FlagOptions {
	var Flags = new(FlagOptions)

	flag.StringVar(&Flags.File, "f", "settings.yaml", "指定配置文件路径")
	flag.BoolVar(&Flags.DB, "db", false, "数据库迁移")
	flag.BoolVar(&Flags.Version, "version", false, "显示版本信息")
	flag.StringVar(&Flags.Type, "t", "", "操作类型")
	flag.StringVar(&Flags.Sub, "s", "", "子操作类型")
	flag.BoolVar(&Flags.ES, "es", false, "初始化ES索引")

	flag.Parse()

	return Flags
}

func Run(op *FlagOptions, db *gorm.DB) {
	if op.DB {
		// 执行数据库迁移
		FlagDB(db)
		os.Exit(0)
	}

	if op.ES {
		switch op.Sub {
		case "init":
			FlagESIndex()
			os.Exit(0)
		case "article-sync":
			FlagESArticleSync()
			os.Exit(0)
		}
		fmt.Println("未知子操作类型")
		os.Exit(0)
	}

	switch op.Type {
	case "user":
		u := FlagUser{}
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
