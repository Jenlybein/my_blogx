package user_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/store/email_store"
	"myblogx/utils/jwts"
	"myblogx/utils/pwd"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RegisterEmailRequest struct {
	EmailID   string `json:"email_id" binding:"required"`
	EmailCode string `json:"email_code" binding:"required"`
	Pwd       string `json:"pwd" binding:"required"`
}

func (u *UserApi) RegisterEmailView(c *gin.Context) {
	var cr RegisterEmailRequest
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	emailInfo, ok := global.EmailVerifyStore.Load(cr.EmailID)
	if !ok {
		res.FailWithMsg("邮箱验证失败：验证码不存在", c)
		return
	}
	info, ok := emailInfo.(*email_store.EmailStoreInfo)
	if !ok {
		res.FailWithMsg("邮箱验证失败：验证码信息获取失败", c)
		return
	}

	if info.Code != cr.EmailCode {
		// TODO：应该有一个超时时间
		// TODO2：最好放在Redis中，而不是本地实现
		info.FailCount++
		global.EmailVerifyStore.Store(cr.EmailID, info) // 更新失败次数
		if info.FailCount >= 3 {
			res.FailWithMsg("邮箱验证码错误次数超过 3 次", c)
			return
		}
		res.FailWithMsg("邮箱验证码错误", c)
		return
	}
	global.EmailVerifyStore.Delete(cr.EmailID) // 成功后删除验证码

	// 注册用户
	hashedPassword, err := pwd.GenerateFromPassword(cr.Pwd)
	if err != nil {
		res.FailWithMsg("邮箱注册失败", c)
		return
	}
	var maxID uint64
	global.DB.Model(&models.UserModel{}).Select("MAX(id)").Scan(&maxID)

	user := models.UserModel{
		Username:       fmt.Sprintf("%d", maxID+1+10000),
		Password:       hashedPassword,
		Nickname:       cr.EmailID,
		Avatar:         "xxx.png",
		RegisterSource: enum.RegisterEmailSourceType,
		Email:          info.Email,
		Role:           enum.RoleUser,
	}
	if err = global.DB.Create(&user).Error; err != nil {
		res.FailWithMsg("邮箱注册失败", c)
		logrus.Errorf("邮箱注册失败 %v", err)
		return
	}

	// 签发token
	jwtToken, err := jwts.GetToken(jwts.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		res.FailWithMsg("邮箱登录失败", c)
		return
	}

	// 返回token
	res.OkWithData(jwtToken, c)
}
