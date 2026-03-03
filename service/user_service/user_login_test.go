package user_service_test

import (
	"myblogx/models"
	"myblogx/service/user_service"
	"myblogx/test/testutil"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserLoginCreateLog(t *testing.T) {
	db := testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{}, &models.UserLoginModel{})

	user := models.UserModel{
		Username: "login_u1",
		Password: "hashed",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	svc := user_service.NewUserService(user)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "127.0.0.1:4567"
	req.Header.Set("User-Agent", "unit-test-agent")
	c.Request = req

	svc.UserLogin(c)

	var log models.UserLoginModel
	if err := db.Last(&log).Error; err != nil {
		t.Fatalf("查询登录日志失败: %v", err)
	}
	if log.UserID != user.ID {
		t.Fatalf("UserID 错误: got=%d want=%d", log.UserID, user.ID)
	}
	if log.IP == "" || log.UA != "unit-test-agent" {
		t.Fatalf("IP/UA 记录异常: %+v", log)
	}
	if log.Addr == "" {
		t.Fatalf("地址字段不应为空: %+v", log)
	}
}
