package redis_user

import (
	"context"
	"fmt"
	"myblogx/global"
)

// keyExists 检查Redis Key是否存在
func keyExists(ctx context.Context, key string) bool {
	exists, err := global.Redis.Exists(ctx, key).Result()
	return err == nil && exists > 0
}

// emailSendKeyByEmail 生成邮箱维度发送频率计数的Redis Key
func emailSendKeyByEmail(email string, sendType int8) string {
	return fmt.Sprintf("auth:email:send:email:%d:%s", sendType, email)
}

// emailSendKeyByIP 生成IP维度发送频率计数的Redis Key
func emailSendKeyByIP(ip string, sendType int8) string {
	return fmt.Sprintf("auth:email:send:ip:%d:%s", sendType, ip)
}

// loginFailUserKey 生成用户登录失败计数的Redis Key
func loginFailUserKey(account string) string {
	return fmt.Sprintf("auth:login:fail:user:%s", account)
}

// loginFailIPKey 生成IP登录失败计数的Redis Key
func loginFailIPKey(ip string) string {
	return fmt.Sprintf("auth:login:fail:ip:%s", ip)
}

// loginLockUserKey 生成用户登录锁定的Redis Key
func loginLockUserKey(account string) string {
	return fmt.Sprintf("auth:login:lock:user:%s", account)
}

// loginLockIPKey 生成IP登录锁定的Redis Key
func loginLockIPKey(ip string) string {
	return fmt.Sprintf("auth:login:lock:ip:%s", ip)
}
