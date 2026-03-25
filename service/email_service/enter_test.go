package email_service_test

import (
	"myblogx/conf"
	"myblogx/global"
	"myblogx/service/email_service"
	"myblogx/test/testutil"
	"testing"
)

func TestSendEmailFunctions(t *testing.T) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		Email: conf.Email{
			Domain:       "",
			Port:         0,
			SendEmail:    "noreply@example.com",
			AuthCode:     "x",
			SendNickname: "BlogX",
		},
	}

	if err := email_service.SendRegisterCode("u@example.com", "1234", 5); err != nil {
		t.Fatalf("SendRegisterCode 返回错误: %v", err)
	}
	if err := email_service.SendResetPwdCode("u@example.com", "1234", 5); err != nil {
		t.Fatalf("SendResetPwdCode 返回错误: %v", err)
	}
	if err := email_service.SendBindEmailCode("u@example.com", "1234", 5); err != nil {
		t.Fatalf("SendBindEmailCode 返回错误: %v", err)
	}
	if err := email_service.SendLoginCode("u@example.com", "1234", 5); err != nil {
		t.Fatalf("SendLoginCode 返回错误: %v", err)
	}
}
