package enum

type RoleType int8

const (
	RoleAdmin RoleType = 1
	RoleUser  RoleType = 2
	RoleGuest RoleType = 3
)
