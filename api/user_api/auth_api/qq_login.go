package auth_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/qq_service"
	"myblogx/service/redis_service/redis_user"
	"myblogx/service/user_service"

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
			username, usernameErr := redis_user.NextAutoUsername()
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

	if !user.CanLogin() {
		res.FailWithMsg(user.Status.String(), c)
		return
	}

	token, refreshToken, _, err := user_service.CreateLoginTokens(&user, user_service.BuildSessionMetaFromGin(c))
	if err != nil {
		res.FailWithMsg("qq登录失败 "+err.Error(), c)
		return
	}
	user_service.SetRefreshTokenCookie(c, refreshToken)
	// 登录日志
	user_service.UserLoginLog(c, user.ID)

	res.OkWithData(token, c)
}
