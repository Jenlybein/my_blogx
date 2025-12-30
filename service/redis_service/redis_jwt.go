package redis_service

import (
	"context"
	"fmt"
	"myblogx/global"
	"myblogx/utils/jwts"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type BlackType int8

const (
	UserBlackType   BlackType = 1 // 用户注销登录
	AdminBlackType  BlackType = 2 // 管理员注销登录
	DeviceBlackType BlackType = 3 // 其他设备已登录
)

// redis 不能自动转化自定义类型，需要手动转换
func (b BlackType) String() string {
	return fmt.Sprintf("%d", b)
}
func BlackTypeFromString(str string) (BlackType, error) {
	num1, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return BlackType(num1), nil
}
func BlackTypeMsg(blackType BlackType) string {
	switch blackType {
	case UserBlackType:
		return "用户注销登录"
	case AdminBlackType:
		return "管理员注销登录"
	case DeviceBlackType:
		return "其他设备已登录"
	default:
		return "未知错误"
	}
}

func TokenBlackList(token string, blackType BlackType) {
	key := fmt.Sprintf("token_blacklist_%s", token)

	// 获取 token 原本的过期时间
	claims, err := jwts.ParseToken(token)
	if err != nil || claims == nil {
		logrus.Errorf("将Token放入黑名单时解析失败 err: %v", err)
		return
	}

	// 计算 token 剩余过期时间
	expire := claims.ExpiresAt - time.Now().Unix()
	if expire <= 0 {
		logrus.Errorf("token 已过期，无法放入黑名单")
		return
	}

	_, err = global.Redis.Set(context.Background(), key, blackType.String(), time.Duration(expire)*time.Second).Result()
	if err != nil {
		logrus.Errorf("将Token放入黑名单时出错 err: %v", err)
		return
	}
}

func HasTokenBlack(token string) (BlackMsg string, ok bool) {
	key := fmt.Sprintf("token_blacklist_%s", token)
	has, err := global.Redis.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return "Token 不在黑名单中", true
		}
		logrus.Errorf("检查Token是否在黑名单时出错 err: %v", err)
		return BlackTypeMsg(0), false
	}

	blackType, err := BlackTypeFromString(has)
	if err != nil {
		logrus.Errorf("string 转换 BlackType 失败: %v", err)
		return BlackTypeMsg(0), false
	}

	return BlackTypeMsg(blackType), true
}

func HasTokenBlackByGin(c *gin.Context) (BlackMsg string, ok bool) {
	token := c.GetHeader("token")
	if token == "" {
		token = c.Query("token")
	}
	return HasTokenBlack(token)
}
