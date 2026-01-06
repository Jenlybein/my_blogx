package enum

type RoleType int8

const (
	RoleAdmin RoleType = 1 // 管理员
	RoleUser  RoleType = 2 // 普通用户
	RoleGuest RoleType = 3 // 访客
)
