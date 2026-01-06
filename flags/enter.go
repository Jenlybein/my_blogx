// flags/enter.go
package flags

//
//

import (
	"flag"
	"fmt"
	"myblogx/flags/flag_user"
	"os"
)

type Options struct {
	/* 定义命令行参数选项结构体,用于存储和处理命令行传入的各种标志参数 */
	File    string
	DB      bool
	Version bool
	Type    string
	Sub     string
}

var FlagOptions = new(Options)

func Parse() {
	flag.StringVar(&FlagOptions.File, "f", "settings.yaml", "指定配置文件路径")
	flag.BoolVar(&FlagOptions.DB, "db", false, "数据库迁移")
	flag.BoolVar(&FlagOptions.Version, "version", false, "显示版本信息")
	flag.StringVar(&FlagOptions.Type, "t", "", "操作类型")
	flag.StringVar(&FlagOptions.Sub, "s", "", "子操作类型")
	flag.Parse()
}

func Run() {
	if FlagOptions.DB {
		// 执行数据库迁移
		FlagDB()
		os.Exit(0)
	}

	switch FlagOptions.Type {
	case "user":
		u := flag_user.FlagUser{}
		switch FlagOptions.Sub {
		case "create":
			u.Create()
			os.Exit(0)
		default:
			fmt.Println("未知子操作类型")
			os.Exit(1)
		}
	}
}
