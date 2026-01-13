package profile_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

type UserBaseInfoResponse struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	CodeAge     int    `json:"code_age"`
	Avatar      string `gorm:"size:256" json:"avatar"`
	Nickname    string `gorm:"size:32" json:"nickname"`
	ViewCount   int    `json:"view_count"`
	FansCount   int    `json:"fans_count"`
	FollowCount int    `json:"follow_count"`
	Place       string `json:"place"`
}

func (ProfileApi) UserBaseInfoView(c *gin.Context) {
	cr := middleware.GetBindQuery[models.IDRequest](c)

	var user models.UserModel
	if err := global.DB.Take(&user, cr.ID).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	data := UserBaseInfoResponse{
		ID:          user.ID,
		CodeAge:     user.CodeAge(),
		Avatar:      user.Avatar,
		Nickname:    user.Nickname,
		ViewCount:   1,
		FansCount:   1,
		FollowCount: 1,
		// ViewCount:   user.UserConfModel.ViewCount,
		// FansCount:   user.UserConfModel.FansCount,
		// FollowCount: user.UserConfModel.FollowCount,
		Place: user.Addr,
	}

	res.OkWithData(data, c)
}
