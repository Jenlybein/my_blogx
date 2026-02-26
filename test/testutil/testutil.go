package testutil

import (
	"fmt"
	"io"
	"myblogx/conf"
	"myblogx/global"
	"net/http"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitGlobals() {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	global.Logger = logger

	if global.Config == nil {
		global.Config = &conf.Config{}
	}
}

func SetupMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	InitGlobals()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("启动 miniredis 失败: %v", err)
	}

	global.Redis = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	t.Cleanup(func() {
		_ = global.Redis.Close()
		mr.Close()
	})

	return mr
}

func SetupSQLite(t *testing.T, models ...any) *gorm.DB {
	t.Helper()
	InitGlobals()

	dsn := fmt.Sprintf("file:%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 sqlite 失败: %v", err)
	}

	if len(models) > 0 {
		if err = db.AutoMigrate(models...); err != nil {
			t.Fatalf("自动迁移失败: %v", err)
		}
	}

	global.DB = db
	return db
}

func NewJSONRequest(method, target, body string) *http.Request {
	req, _ := http.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
