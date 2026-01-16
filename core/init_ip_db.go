// IP数据库初始化

package core

import (
	"fmt"
	"strings"

	"myblogx/global"
	ipUtils "myblogx/utils/ip"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

var (
	ipv4Searcher *xdb.Searcher
	ipv6Searcher *xdb.Searcher
)

func InitIPDB() {
	var dbIPv4 = "init/ipbase/ip2region_v4.xdb"
	var dbIPv6 = "init/ipbase/ip2region_v6.xdb"

	// 初始化 IPv4 数据库
	ipv4, err := xdb.NewWithFileOnly(xdb.IPv4, dbIPv4)
	if err != nil {
		global.Logger.Fatalf("ip2region_v4.xdb 加载失败: %s", err)
	}
	ipv4Searcher = ipv4

	// 初始化 IPv6 数据库
	ipv6, err := xdb.NewWithFileOnly(xdb.IPv6, dbIPv6)
	if err != nil {
		global.Logger.Fatalf("ip2region_v6.xdb 加载失败: %s", err)
	}
	ipv6Searcher = ipv6
}

func GetIpAddr(ip string) (addr string) {
	// 解析 IP 地址
	ipType := ipUtils.IpType(ip)
	if ipType == "1" {
		return "未知地址"
	}

	if ipUtils.IsLocalIP(ip, ipType) {
		return "内网地址"
	}

	// 判断是 ipv4 还是 ipv6，根据对应搜索器查询
	var region string
	var err error

	switch ipType {
	case "ipv4":
		region, err = ipv4Searcher.SearchByStr(ip)
	case "ipv6":
		region, err = ipv6Searcher.SearchByStr(ip)
	}

	if err != nil || region == "" {
		global.Logger.Warnf("IP 地址 %s 区域查询失败", ip)
		return "未知地址"
	}

	_addrList := strings.Split(region, "|")
	if len(_addrList) < 4 {
		global.Logger.Warnf("IP 地址 %s 区域查询结果格式错误", ip)
		return "未知地址"
	}

	// 提取国家、省份、城市、运营商
	country := _addrList[0]
	province := _addrList[1]
	city := _addrList[2]

	if country == "中国" {
		if province != "0" && city != "0" {
			return fmt.Sprintf("%s·%s", province, city)
		}
	} else {
		if country != "0" && province != "0" {
			return fmt.Sprintf("%s·%s", country, province)
		}
		if country != "0" {
			return country
		}
	}

	return "未知地址"
}
