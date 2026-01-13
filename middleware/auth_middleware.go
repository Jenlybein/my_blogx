package middleware

import (
	"myblogx/common/res"
	"myblogx/models/enum"
	"myblogx/service/redis_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	// 从请求头中获取 token
	claims, err := jwts.ParseTokenByGin(c)
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}

	// 判断 token 是否在黑名单中
	blackMsg, ok := redis_service.HasTokenBlackByGin(c)
	if !ok {
		res.FailWithMsg(blackMsg, c)
		c.Abort()
		return
	}
	c.Set("claims", claims)
}

func AdminMiddleware(c *gin.Context) {
	claims := jwts.GetClaimsByGin(c)

	if claims.Role != enum.RoleAdmin {
		res.FailWithMsg("权限错误", c)
		c.Abort()
		return
	}
}
