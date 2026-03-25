package auth_api

import (
	"myblogx/common/res"
	"myblogx/service/user_service"

	"github.com/gin-gonic/gin"
)

func (AuthApi) RefreshTokenView(c *gin.Context) {
	// 获取旧的刷新令牌
	refreshToken := user_service.GetRefreshTokenByGin(c)

	// 用旧的刷新令牌换取新的AccessToken和新的刷新令牌
	accessToken, newRefreshToken, _, _, err := user_service.RefreshTokens(refreshToken, user_service.BuildSessionMetaFromGin(c))
	
	if err != nil {
		user_service.ClearRefreshTokenCookie(c)
		res.FailWithMsg(err.Error(), c)
		return
	}
	user_service.SetRefreshTokenCookie(c, newRefreshToken)

	res.OkWithData(accessToken, c)
}
