package user_info

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// 获取客户端 IP 地址
func GetClientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		ipList := strings.Split(ip, ",")
		clientIP := strings.TrimSpace(ipList[0])
		if isValidIP(clientIP) {
			return clientIP
		}
	}

	ip = c.GetHeader("X-Real-IP")
	if ip != "" && isValidIP(ip) {
		return ip
	}

	ip = c.GetHeader("CF-Connecting-IP")
	if ip != "" && isValidIP(ip) {
		return ip
	}

	return c.ClientIP()
}

// 返回 IP 地址的类型，包括 ipv4、ipv6 和 空字符串
func IpType(ipstr string) string {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return ""
	}
	if ip.To4() != nil {
		return "ipv4"
	}
	if ip.To16() != nil {
		return "ipv6"
	}
	return ""
}

// 判断 IP 地址是否为本地回环地址
func IsLocalIP(ipstr string, ipType string) bool {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return false
	}

	// 回环地址（127.0.0.0/8 或 ::1）
	if ip.IsLoopback() {
		return true
	}

	// 根据 IP 类型判断私有地址
	switch ipType {
	case "ipv4":
		// 私有地址范围：10.0.0.0/8、172.16.0.0/12、192.168.0.0/16
		return ip[0] == 10 ||
			(ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31) ||
			(ip[0] == 192 && ip[1] == 168)

	case "ipv6":
		// 私有地址范围：链路本地（fe80::/10）、唯一本地（fc00::/7）
		return ip.IsLinkLocalUnicast() || // 链路本地地址（fe80::/10）
			(ip[0]&0x80 == 0x80 && ip[0]&0x40 == 0x40) // 唯一本地地址（fc00::/7）
	}

	return false
}
