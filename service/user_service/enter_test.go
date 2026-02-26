package user_service_test

import (
	"myblogx/models"
	"myblogx/service/user_service"
	"testing"
)

func TestNewUserService(t *testing.T) {
	user := models.UserModel{
		Username: "u1",
	}
	s := user_service.NewUserService(user)
	if s == nil {
		t.Fatal("NewUserService 不应返回 nil")
	}
}
