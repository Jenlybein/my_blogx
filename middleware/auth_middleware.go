package middleware

import (
	"myblogx/common/res"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	authResult, err := user_service.AuthenticateAccessTokenByGin(c)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		c.Abort()
		return
	}

	c.Set("claims", authResult.Claims)
	c.Set("auth_user", authResult.User)
	c.Set("auth_session", authResult.Session)
	c.Set("access_token", authResult.Token)
	c.Next()
}

func AdminMiddleware(c *gin.Context) {
	claims := jwts.MustGetClaimsByGin(c)

	if !claims.IsAdmin() {
		res.FailWithMsg("权限错误", c)
		c.Abort()
		return
	}
	c.Next()
}
