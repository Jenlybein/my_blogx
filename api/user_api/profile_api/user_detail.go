package profile_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

type UserDetailResponse struct {
	ID             ctype.ID                `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time               `json:"created_at"`
	Username       string                  `gorm:"size:32" json:"username"`
	Nickname       string                  `gorm:"size:32" json:"nickname"`
	Avatar         string                  `gorm:"size:256" json:"avatar"`
	Abstract       string                  `gorm:"size:256" json:"abstract"`
	Email          *string                 `json:"email"`
	HasPassword    bool                    `json:"has_password"`
	RegisterSource enum.RegisterSourceType `json:"register_source"`
	CodeAge        int                     `json:"code_age"`
	models.UserConfModel
}

func (ProfileApi) UserDetailView(c *gin.Context) {
	claims := jwts.MustGetClaimsByGin(c)

	var user models.UserModel
	if err := global.DB.Preload("UserConfModel").Take(&user, claims.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		c.Abort()
		return
	}

	var data = UserDetailResponse{
		ID:             user.ID,
		CreatedAt:      user.CreatedAt,
		Username:       user.Username,
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		Abstract:       user.Abstract,
		Email:          user.Email,
		HasPassword:    user.Password != "",
		RegisterSource: user.RegisterSource,
		CodeAge:        user.CodeAge(),
	}
	if user.UserConfModel != nil {
		data.UserConfModel = *user.UserConfModel
	}

	res.OkWithData(data, c)
}
