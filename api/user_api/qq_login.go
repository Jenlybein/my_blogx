package user_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/qq_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

type QQLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

func (u *UserApi) QQLoginView(c *gin.Context) {
	var cr QQLoginRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, c)
		return
	}

	userInfoResp, err := qq_service.GetUserInfo(cr.Code)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var user models.UserModel
	if err = global.DB.Take(&user, "open_id = ?", userInfoResp.OpenID).Error; err != nil {
		// 创建用户
		var maxID uint64
		global.DB.Model(&models.UserModel{}).Select("MAX(id)").Scan(&maxID)
		user = models.UserModel{
			Username:       fmt.Sprintf("%d", maxID+1+10000),
			Nickname:       userInfoResp.NickName,
			Avatar:         userInfoResp.Avatar,
			RegisterSource: enum.RegisterQQSourceType,
			OpenID:         userInfoResp.OpenID,
			Role:           enum.RoleUser,
		}
		if err = global.DB.Create(&user).Error; err != nil {
			res.FailWithMsg("qq登录失败 "+err.Error(), c)
			return
		}
	}

	// 签发token
	token, err := jwts.GetToken(jwts.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	res.OkWithData(token, c)
}
