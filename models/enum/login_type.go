// 登录类型枚举

package enum

type LoginType int8

const (
	PasswordLoginType LoginType = 1 // 密码登录
	QQLoginType       LoginType = 2 // QQ登录
	EmailLoginType    LoginType = 3 // 邮箱登录
)
