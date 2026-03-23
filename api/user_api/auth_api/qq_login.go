package auth_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/qq_service"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QQLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

func (AuthApi) QQLoginView(c *gin.Context) {
	if !global.Config.Site.Login.QQLogin {
		res.FailWithMsg("站点未启用qq登录功能", c)
		return
	}

	cr := middleware.GetBindJson[QQLoginRequest](c)

	userInfoResp, err := qq_service.GetUserInfo(cr.Code)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var user models.UserModel
	err = global.DB.Take(&user, "open_id = ?", userInfoResp.OpenID).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			res.FailWithMsg("qq登录失败 "+err.Error(), c)
			return
		}

		for range 5 {
			username, usernameErr := user_service.NextAutoUsername()
			if usernameErr != nil {
				global.Logger.Errorf("qq 登录生成用户名失败: %v", usernameErr)
				res.FailWithMsg("qq登录失败", c)
				return
			}

			openID := userInfoResp.OpenID
			user = models.UserModel{
				Username:       username,
				Nickname:       userInfoResp.NickName,
				Avatar:         userInfoResp.Avatar,
				RegisterSource: enum.RegisterQQSourceType,
				OpenID:         &openID,
				Role:           enum.RoleUser,
			}
			result := global.DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "open_id"}},
				DoNothing: true,
			}).Create(&user)
			if result.Error == nil {
				if result.RowsAffected == 0 {
					if err = global.DB.Take(&user, "open_id = ?", userInfoResp.OpenID).Error; err != nil {
						res.FailWithMsg("qq登录失败 "+err.Error(), c)
						return
					}
				}
				break
			}
			if !errors.Is(result.Error, gorm.ErrDuplicatedKey) {
				res.FailWithMsg("qq登录失败 "+result.Error.Error(), c)
				return
			}
		}
		if user.ID == 0 {
			res.FailWithMsg("qq登录失败", c)
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
		res.FailWithMsg("qq登录失败 "+err.Error(), c)
		return
	}
	// 登录日志
	user_service.NewUserService(user).UserLogin(c)

	res.OkWithData(token, c)
}
