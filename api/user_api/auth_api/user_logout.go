package auth_api

import (
	"myblogx/common/res"
	"myblogx/service/redis_service/redis_jwt"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (AuthApi) UserLogoutView(c *gin.Context) {
	claims := jwts.MustGetClaimsByGin(c)
	if err := user_service.RevokeSessionByID(claims.UserID, claims.SessionID); err != nil {
		res.FailWithError(err, c)
		return
	}

	token := jwts.GetTokenByGin(c)
	if token != "" {
		redis_jwt.SetTokenBlack(token, redis_jwt.UserBlackType)
	}
	user_service.ClearRefreshTokenCookie(c)

	res.OkWithMsg("退出登录成功", c)
}

func (AuthApi) UserLogoutAllView(c *gin.Context) {
	claims := jwts.MustGetClaimsByGin(c)
	if err := user_service.RevokeAllUserSessions(claims.UserID); err != nil {
		res.FailWithError(err, c)
		return
	}

	token := jwts.GetTokenByGin(c)
	if token != "" {
		redis_jwt.SetTokenBlack(token, redis_jwt.UserBlackType)
	}
	user_service.ClearRefreshTokenCookie(c)

	res.OkWithMsg("已退出全部设备", c)
}
