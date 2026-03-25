package email_service

import (
	"fmt"
	"myblogx/global"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// 注册账号
func SendRegisterCode(to string, code string, timeout int) error {
	em := global.Config.Email
	subject := fmt.Sprintf("【%s】 账号注册", em.SendNickname)
	text := fmt.Sprintf("您正在进行账号注册操作，这是您的验证码为：%s，有效期为 %d 分钟", code, timeout)
	return SendEmail(to, subject, text)
}

// 重置密码
func SendResetPwdCode(to string, code string, timeout int) error {
	em := global.Config.Email
	subject := fmt.Sprintf("【%s】 重置密码", em.SendNickname)
	text := fmt.Sprintf("您正在进行重置密码操作，这是您的验证码为：%s，有效期为 %d 分钟", code, timeout)
	return SendEmail(to, subject, text)
}

// 绑定邮箱
func SendBindEmailCode(to string, code string, timeout int) error {
	em := global.Config.Email
	subject := fmt.Sprintf("【%s】 绑定邮箱", em.SendNickname)
	text := fmt.Sprintf("您正在进行绑定邮箱操作，这是您的验证码为：%s，有效期为 %d 分钟", code, timeout)
	return SendEmail(to, subject, text)
}

// 邮箱登录
func SendLoginCode(to string, code string, timeout int) error {
	em := global.Config.Email
	subject := fmt.Sprintf("【%s】 邮箱登录", em.SendNickname)
	text := fmt.Sprintf("您正在进行邮箱登录操作，这是您的验证码为：%s，有效期为 %d 分钟", code, timeout)
	return SendEmail(to, subject, text)
}

// 发送邮件
func SendEmail(to string, subject string, text string) (err error) {
	emcfg := global.Config.Email
	e := &email.Email{
		From:    fmt.Sprintf("%s <%s>", emcfg.SendNickname, emcfg.SendEmail),
		To:      []string{to},
		Subject: subject,
		Text:    []byte(text),
	}
	addr := fmt.Sprintf("%s:%d", emcfg.Domain, emcfg.Port)
	if err = e.Send(addr, smtp.PlainAuth("",
		emcfg.SendEmail,
		emcfg.AuthCode,
		emcfg.Domain,
	)); err != nil {
		fmt.Println(e)
	}

	return nil
}
