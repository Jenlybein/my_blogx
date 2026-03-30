package profile_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/service/follow_service"
	"myblogx/service/user_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

type UserBaseInfoResponse struct {
	ID                  ctype.ID `gorm:"primaryKey" json:"id"`
	CodeAge             int      `json:"code_age"`
	Avatar              string   `gorm:"size:256" json:"avatar"`
	Nickname            string   `gorm:"size:32" json:"nickname"`
	ViewCount           int      `json:"view_count"`
	FansCount           int      `json:"fans_count"`
	FollowCount         int      `json:"follow_count"`
	FavoritesVisibility bool     `json:"favorites_visibility"`
	FollowVisibility    bool     `json:"followers_visibility"`
	FansVisibility      bool     `json:"fans_visibility"`
	HomeStyleID         ctype.ID `json:"home_style_id"`
	Relation            int8     `json:"relation"`
	Place               string   `json:"place"`
}

func (ProfileApi) UserBaseInfoView(c *gin.Context) {
	cr := middleware.GetBindQuery[models.IDRequest](c)

	var user models.UserModel
	if err := global.DB.Preload("UserConfModel").Preload("UserStatModel").Take(&user, cr.ID).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	viewCountDelta := 0
	if claims := jwts.GetClaimsByGin(c); claims != nil && claims.UserID != 0 && claims.UserID != user.ID {
		counted, err := user_service.StatRecordUserHomeView(user.ID, claims.UserID)
		if err != nil {
			res.FailWithError(err, c)
			return
		}
		if counted {
			viewCountDelta = 1
		}
	}

	var stat models.UserStatModel
	if user.UserStatModel != nil {
		stat = *user.UserStatModel
	}
	var conf models.UserConfModel
	if user.UserConfModel != nil {
		conf = *user.UserConfModel
	}
	relation := int8(0)
	if claims := jwts.GetClaimsByGin(c); claims != nil {
		relation = int8(follow_service.CalUserRelationship(claims.UserID, user.ID))
	}

	data := UserBaseInfoResponse{
		ID:                  user.ID,
		CodeAge:             user.CodeAge(),
		Avatar:              user.Avatar,
		Nickname:            user.Nickname,
		ViewCount:           stat.ViewCount + viewCountDelta,
		FansCount:           stat.FansCount,
		FollowCount:         stat.FollowCount,
		FavoritesVisibility: conf.FavoritesVisibility,
		FollowVisibility:    conf.FollowVisibility,
		FansVisibility:      conf.FansVisibility,
		HomeStyleID:         conf.HomeStyleID,
		Relation:            relation,
		Place:               user.Addr,
	}

	res.OkWithData(data, c)
}
