package auth_api

import (
	"myblogx/common/res"
	"myblogx/service/redis_service/redis_jwt"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (AuthApi) UserLogoutView(c *gin.Context) {
	token := jwts.GetTokenByGin(c)
	redis_jwt.SetTokenBlack(token, redis_jwt.UserBlackType)

	res.OkWithMsg("退出登录成功", c)
}
