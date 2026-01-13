package profile_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/info_check"
	"myblogx/utils/jwts"
	"myblogx/utils/maps"
	"time"

	"github.com/gin-gonic/gin"
)

type UserInfoUpdateRequest struct {
	Username            *string   `json:"username"`
	Nickname            *string   `json:"nickname"`
	Avatar              *string   `json:"avatar"`
	Abstract            *string   `json:"abstract"`
	LikeTags            *[]string `json:"like_tags"`
	FavoritesVisibility *bool     `json:"favorites_visibility"`
	FollowVisibility    *bool     `json:"followers_visibility"`
	FansVisibility      *bool     `json:"fans_visibility"`
	HomeStyleID         *uint     `json:"home_style_id"`
}

func (ProfileApi) UserInfoUpdateView(c *gin.Context) {
	cr := middleware.GetBindJson[UserInfoUpdateRequest](c)

	userMap, err := maps.FieldsStructToMap(&cr, &models.UserModel{})
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	confMap, err := maps.FieldsStructToMap(&cr, &models.UserConfModel{})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	claims := jwts.GetClaimsByGin(c)

	// 处理用户基本表的更新
	if len(userMap) > 0 {
		var userModel models.UserModel
		if err = global.DB.Preload("UserConfModel").Take(&userModel, claims.UserID).Error; err != nil {
			res.FailWithMsg("用户不存在", c)
			return
		}

		if cr.Username != nil && *cr.Username != userModel.Username {
			// 校验用户名格式
			if err = info_check.CheckUsername(*cr.Username); err != nil {
				res.FailWithError(err, c)
				return
			}

			// 校验用户名是否已被使用
			var nameCount int64
			if err = global.DB.Model(&models.UserModel{}).Where("username = ?", *cr.Username).Count(&nameCount).Error; err != nil {
				res.FailWithError(err, c)
				return
			}
			if nameCount > 0 {
				res.FailWithMsg("用户名已被使用", c)
				return
			}

			// 校验用户名更新频率
			uud := userModel.UserConfModel.UpdatedUsernameDate
			updateLimit := time.Hour * 720
			if time.Since(uud) < updateLimit {
				res.FailWithMsg(fmt.Sprintf("用户名每 %d 天内只能更新 1 次", int(updateLimit.Hours()/24)), c)
				return
			}

			confMap["updated_username_date"] = time.Now()
		}

		if err = global.DB.Model(&userModel).Updates(userMap).Error; err != nil {
			res.FailWithMsg("用户信息更新失败", c)
			return
		}
	}

	// 处理用户配置表的更新
	if len(confMap) > 0 {
		var userConfModel models.UserConfModel
		if err = global.DB.Take(&userConfModel, claims.UserID).Error; err != nil {
			res.FailWithMsg("用户配置信息不存在", c)
			return
		}

		if err = global.DB.Model(&userConfModel).Updates(confMap).Error; err != nil {
			res.FailWithMsg("用户配置信息更新失败", c)
			return
		}
	}

	res.OkWithMsg("用户信息更新成功", c)
}
