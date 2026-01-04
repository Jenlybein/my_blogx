package enum

type RoleType int8

// 角色 0:管理员 1:普通用户 2:访客
const (
	RoleAdmin RoleType = 1
	RoleUser  RoleType = 2
	RoleGuest RoleType = 3
)
