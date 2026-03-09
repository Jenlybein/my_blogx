package testutil

import (
	"fmt"
	"io"
	"myblogx/conf"
	"myblogx/global"
	blogmodels "myblogx/models"
	"net/http"
	"reflect"
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

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取 sqlite 连接失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if len(models) > 0 {
		models = appendArticleTagModels(models)
		if err = db.AutoMigrate(models...); err != nil {
			t.Fatalf("自动迁移失败: %v", err)
		}
	}

	global.DB = db
	return db
}

func appendArticleTagModels(list []any) []any {
	hasArticle := false
	hasTag := false
	hasArticleTag := false

	for _, item := range list {
		switch reflect.TypeOf(item) {
		case reflect.TypeOf(&blogmodels.ArticleModel{}):
			hasArticle = true
		case reflect.TypeOf(&blogmodels.TagModel{}):
			hasTag = true
		case reflect.TypeOf(&blogmodels.ArticleTagModel{}):
			hasArticleTag = true
		}
	}

	if !hasArticle {
		return list
	}
	if !hasTag {
		list = append(list, &blogmodels.TagModel{})
	}
	if !hasArticleTag {
		list = append(list, &blogmodels.ArticleTagModel{})
	}
	return list
}

func NewJSONRequest(method, target, body string) *http.Request {
	req, _ := http.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
