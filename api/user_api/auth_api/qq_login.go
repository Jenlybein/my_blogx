package auth_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/log_service"
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
		log_service.EmitLoginEventFromGin(c, "login_fail", enum.QQLoginType, false, "", 0, "站点未启用QQ登录", nil)
		res.FailWithMsg("站点未启用qq登录功能", c)
		return
	}

	cr := middleware.GetBindJson[QQLoginRequest](c)

	userInfoResp, err := qq_service.GetUserInfo(cr.Code)
	if err != nil {
		log_service.EmitLoginEventFromGin(c, "login_fail", enum.QQLoginType, false, "", 0, err.Error(), nil)
		res.FailWithError(err, c)
		return
	}

	var user models.UserModel
	err = global.DB.Take(&user, "open_id = ?", userInfoResp.OpenID).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log_service.EmitLoginEventFromGin(c, "login_fail", enum.QQLoginType, false, userInfoResp.NickName, 0, "qq登录失败 "+err.Error(), map[string]any{
				"open_id": userInfoResp.OpenID,
			})
			res.FailWithMsg("qq登录失败 "+err.Error(), c)
			return
		}

		for range 5 {
			username, usernameErr := redis_user.NextAutoUsername()
			if usernameErr != nil {
				global.Logger.Errorf("QQ 登录生成用户名失败: %v", usernameErr)
				log_service.EmitLoginEventFromGin(c, "login_fail", enum.QQLoginType, false, userInfoResp.NickName, 0, "qq登录失败", map[string]any{
					"open_id": userInfoResp.OpenID,
				})
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
						log_service.EmitLoginEventFromGin(c, "login_fail", enum.QQLoginType, false, userInfoResp.NickName, 0, "qq登录失败 "+err.Error(), map[string]any{
							"open_id": userInfoResp.OpenID,
						})
						res.FailWithMsg("qq登录失败 "+err.Error(), c)
						return
					}
				}
				break
			}
			if !errors.Is(result.Error, gorm.ErrDuplicatedKey) {
				log_service.EmitLoginEventFromGin(c, "login_fail", enum.QQLoginType, false, userInfoResp.NickName, 0, "qq登录失败 "+result.Error.Error(), map[string]any{
					"open_id": userInfoResp.OpenID,
				})
				res.FailWithMsg("qq登录失败 "+result.Error.Error(), c)
				return
			}
		}
		if user.ID == 0 {
			log_service.EmitLoginEventFromGin(c, "login_fail", enum.QQLoginType, false, userInfoResp.NickName, 0, "qq登录失败", map[string]any{
				"open_id": userInfoResp.OpenID,
			})
			res.FailWithMsg("qq登录失败", c)
			return
		}
	}

	if !user.CanLogin() {
		log_service.EmitLoginEventFromGin(c, "login_fail", enum.QQLoginType, false, user.Username, user.ID, user.Status.String(), map[string]any{
			"open_id": userInfoResp.OpenID,
		})
		res.FailWithMsg(user.Status.String(), c)
		return
	}

	token, refreshToken, _, err := user_service.CreateLoginTokens(&user, user_service.BuildSessionMetaFromGin(c))
	if err != nil {
		log_service.EmitLoginEventFromGin(c, "login_fail", enum.QQLoginType, false, user.Username, user.ID, "qq登录失败 "+err.Error(), map[string]any{
			"open_id": userInfoResp.OpenID,
		})
		res.FailWithMsg("qq登录失败 "+err.Error(), c)
		return
	}
	user_service.SetRefreshTokenCookie(c, refreshToken)
	log_service.EmitLoginEventFromGin(c, "login_success", enum.QQLoginType, true, user.Username, user.ID, "", map[string]any{
		"open_id": userInfoResp.OpenID,
	})

	res.OkWithData(token, c)
}
