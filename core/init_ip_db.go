// IP数据库初始化

package core

import (
	"myblogx/global"
	"myblogx/utils/ipmeta"
)

func InitIPDB() {
	var dbIPv4 = "init/ipbase/ip2region_v4.xdb"
	var dbIPv6 = "init/ipbase/ip2region_v6.xdb"

	if err := ipmeta.Init(dbIPv4, dbIPv6); err != nil {
		global.Logger.Fatalf("IP数据库初始化失败:%s", err)
	}
}

func GetIpAddr(ip string) (addr string) {
	return ipmeta.GetAddr(ip)
}
