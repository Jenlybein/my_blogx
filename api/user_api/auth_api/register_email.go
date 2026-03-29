package auth_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/log_service"
	"myblogx/service/redis_service/redis_user"
	"myblogx/service/user_service"
	"myblogx/utils/pwd"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RegisterEmailRequest struct {
	Pwd string `json:"pwd" binding:"required"`
}

func (AuthApi) RegisterEmailView(c *gin.Context) {
	if !global.Config.Site.Login.EmailLogin {
		log_service.EmitLoginEventFromGin(c, "register_fail", enum.EmailLoginType, false, "", 0, "站点未启用邮箱注册", nil)
		res.FailWithMsg("站点未启用邮箱注册功能", c)
		return
	}

	cr := middleware.GetBindJson[RegisterEmailRequest](c)

	email := c.GetString("email")
	if email == "" {
		log_service.EmitLoginEventFromGin(c, "register_fail", enum.EmailLoginType, false, "", 0, "邮箱验证失败：邮箱不存在", nil)
		res.FailWithMsg("邮箱验证失败：邮箱不存在", c)
		return
	}

	// 注册用户
	hashedPassword, err := pwd.GenerateFromPassword(cr.Pwd)
	if err != nil {
		log_service.EmitLoginEventFromGin(c, "register_fail", enum.EmailLoginType, false, email, 0, "邮箱注册失败", nil)
		res.FailWithMsg("邮箱注册失败", c)
		return
	}
	username, err := redis_user.NextAutoUsername()
	if err != nil {
		global.Logger.Errorf("邮箱注册生成用户名失败: %v", err)
		log_service.EmitLoginEventFromGin(c, "register_fail", enum.EmailLoginType, false, email, 0, "邮箱注册失败", nil)
		res.FailWithMsg("邮箱注册失败", c)
		return
	}

	var user models.UserModel
	for range 5 {
		emailValue := email
		user = models.UserModel{
			Username:       username,
			Password:       hashedPassword,
			Nickname:       email,
			Avatar:         "xxx.png",
			RegisterSource: enum.RegisterEmailSourceType,
			Email:          &emailValue,
			Role:           enum.RoleUser,
		}
		result := global.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "email"}},
			DoNothing: true,
		}).Create(&user)
		if result.Error == nil {
			if result.RowsAffected == 0 {
				log_service.EmitLoginEventFromGin(c, "register_fail", enum.EmailLoginType, false, email, 0, "邮箱已被使用", nil)
				res.FailWithMsg("邮箱已被使用", c)
				return
			}
			break
		}
		if !errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			err = result.Error
			break
		}

		username, err = redis_user.NextAutoUsername()
		if err != nil {
			global.Logger.Errorf("邮箱注册生成用户名失败: %v", err)
			log_service.EmitLoginEventFromGin(c, "register_fail", enum.EmailLoginType, false, email, 0, "邮箱注册失败", nil)
			res.FailWithMsg("邮箱注册失败", c)
			return
		}
	}
	if err != nil {
		log_service.EmitLoginEventFromGin(c, "register_fail", enum.EmailLoginType, false, email, 0, "邮箱注册失败", nil)
		res.FailWithMsg("邮箱注册失败", c)
		global.Logger.Errorf("邮箱注册失败 %v", err)
		return
	}
	if user.ID == 0 {
		log_service.EmitLoginEventFromGin(c, "register_fail", enum.EmailLoginType, false, email, 0, "邮箱注册失败", nil)
		res.FailWithMsg("邮箱注册失败", c)
		global.Logger.Errorf("邮箱注册失败: 自动用户名重试次数耗尽")
		return
	}

	jwtToken, refreshToken, _, err := user_service.CreateLoginTokens(&user, user_service.BuildSessionMetaFromGin(c))
	if err != nil {
		log_service.EmitLoginEventFromGin(c, "register_fail", enum.EmailLoginType, false, user.Username, user.ID, "邮箱登录失败", nil)
		res.FailWithMsg("邮箱登录失败", c)
		return
	}
	user_service.SetRefreshTokenCookie(c, refreshToken)
	log_service.EmitLoginEventFromGin(c, "register_success", enum.EmailLoginType, true, user.Username, user.ID, "", map[string]any{
		"email": email,
	})
	log_service.EmitLoginEventFromGin(c, "login_success", enum.EmailLoginType, true, user.Username, user.ID, "", map[string]any{
		"email": email,
	})

	// 返回token
	res.OkWithData(jwtToken, c)
}
