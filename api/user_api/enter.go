package user_api

import (
	"myblogx/api/user_api/auth_api"
	"myblogx/api/user_api/log_api"
	"myblogx/api/user_api/profile_api"
	"myblogx/api/user_api/user_man_api"
)

type UserApi struct {
	ProfileApi profile_api.ProfileApi
	AuthApi    auth_api.AuthApi
	LogApi     log_api.LogApi
	UserManApi user_man_api.UserManApi
}
