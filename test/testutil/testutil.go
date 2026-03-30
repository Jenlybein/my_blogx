package testutil

import (
	"fmt"
	"io"
	"myblogx/conf"
	"myblogx/global"
	blogmodels "myblogx/models"
	"myblogx/models/ctype"
	"myblogx/service/db_service"
	"myblogx/utils/jwts"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
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
	if global.Config.Log.Dir == "" {
		global.Config.Log.Dir = defaultTestLogDir()
	}
	if global.Config.Log.App == "" {
		global.Config.Log.App = "test"
	}
	if global.Config.Log.StdoutFormat == "" {
		global.Config.Log.StdoutFormat = "json"
	}
	if global.Config.System.ServerID == 0 {
		global.Config.System.ServerID = 1
	}
	if err := db_service.InitSnowflake(global.Config.System.ServerID); err != nil {
		panic(fmt.Errorf("初始化测试雪花 ID 生成器失败: %w", err))
	}
}

func defaultTestLogDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "./logs/test_logs"
	}
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	return filepath.Join(projectRoot, "logs", "test_logs")
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
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
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
		models = appendDependentModels(models)
		if err = db.AutoMigrate(models...); err != nil {
			t.Fatalf("自动迁移失败: %v", err)
		}
	}

	global.DB = db
	return db
}

func appendDependentModels(list []any) []any {
	hasArticle := false
	hasTag := false
	hasArticleTag := false
	hasUser := false
	hasUserConf := false
	hasUserStat := false
	hasUserSession := false

	for _, item := range list {
		switch reflect.TypeOf(item) {
		case reflect.TypeOf(&blogmodels.ArticleModel{}):
			hasArticle = true
		case reflect.TypeOf(&blogmodels.TagModel{}):
			hasTag = true
		case reflect.TypeOf(&blogmodels.ArticleTagModel{}):
			hasArticleTag = true
		case reflect.TypeOf(&blogmodels.UserModel{}):
			hasUser = true
		case reflect.TypeOf(&blogmodels.UserConfModel{}):
			hasUserConf = true
		case reflect.TypeOf(&blogmodels.UserStatModel{}):
			hasUserStat = true
		case reflect.TypeOf(&blogmodels.UserSessionModel{}):
			hasUserSession = true
		}
	}

	if hasUser {
		if !hasUserConf {
			list = append(list, &blogmodels.UserConfModel{})
		}
		if !hasUserStat {
			list = append(list, &blogmodels.UserStatModel{})
		}
		if !hasUserSession {
			list = append(list, &blogmodels.UserSessionModel{})
		}
	}

	if hasArticle {
		if !hasTag {
			list = append(list, &blogmodels.TagModel{})
		}
		if !hasArticleTag {
			list = append(list, &blogmodels.ArticleTagModel{})
		}
	}
	return list
}

func NewJSONRequest(method, target, body string) *http.Request {
	req, _ := http.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func IssueAccessToken(t *testing.T, user *blogmodels.UserModel) string {
	t.Helper()
	sessionID, err := db_service.NextSnowflakeID()
	if err != nil {
		t.Fatalf("生成测试会话ID失败: %v", err)
	}
	now := time.Now()
	session := blogmodels.UserSessionModel{
		Model: blogmodels.Model{
			ID: sessionID,
		},
		UserID:           user.ID,
		RefreshTokenHash: fmt.Sprintf("test-refresh-%s", sessionID.String()),
		IP:               "127.0.0.1",
		Addr:             "本地测试",
		UA:               "codex-test",
		LastSeenAt:       &now,
		ExpiresAt:        now.Add(24 * time.Hour),
	}
	if err = global.DB.Create(&session).Error; err != nil {
		t.Fatalf("创建测试会话失败: %v", err)
	}

	token, err := jwts.GetToken(jwts.Claims{
		UserID:       user.ID,
		SessionID:    session.ID,
		TokenVersion: effectiveTokenVersion(user),
		Username:     user.Username,
		Role:         user.Role,
	})
	if err != nil {
		t.Fatalf("签发测试访问令牌失败: %v", err)
	}
	return token
}

func effectiveTokenVersion(user *blogmodels.UserModel) uint32 {
	if user == nil || user.TokenVersion == 0 {
		return 1
	}
	return user.TokenVersion
}

func NewClaims(user *blogmodels.UserModel) *jwts.MyClaims {
	if user == nil {
		return &jwts.MyClaims{}
	}
	return &jwts.MyClaims{
		Claims: jwts.Claims{
			UserID:       user.ID,
			SessionID:    ctype.ID(0),
			TokenVersion: effectiveTokenVersion(user),
			Username:     user.Username,
			Role:         user.Role,
		},
	}
}
