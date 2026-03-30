package auth_api

import (
	"myblogx/common/res"
	"myblogx/models/enum"
	"myblogx/service/log_service"
	"myblogx/service/user_service"

	"github.com/gin-gonic/gin"
)

func (AuthApi) RefreshTokenView(c *gin.Context) {
	// 获取旧的刷新令牌
	refreshToken := user_service.GetRefreshTokenByGin(c)

	// 用旧的刷新令牌换取新的AccessToken和新的刷新令牌
	accessToken, newRefreshToken, _, _, err := user_service.RefreshTokens(refreshToken, user_service.BuildSessionMetaFromGin(c))

	if err != nil {
		log_service.EmitLoginEventFromGin(c, "token_refresh", enum.LoginType(0), false, "", 0, err.Error(), nil)
		user_service.ClearRefreshTokenCookie(c)
		res.FailWithMsg(err.Error(), c)
		return
	}
	user_service.SetRefreshTokenCookie(c, newRefreshToken)
	log_service.EmitLoginEventFromGin(c, "token_refresh", enum.LoginType(0), true, "", 0, "", nil)

	res.OkWithData(accessToken, c)
}
