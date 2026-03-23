package redis_site

import (
	"context"
	"myblogx/global"
)

func SetFlow() {
	global.Redis.IncrBy(context.Background(), "blog_site_flow", 1).Err()
}

func GetFlow() int {
	num, _ := global.Redis.Get(context.Background(), "blog_site_flow").Int()
	return num
}
