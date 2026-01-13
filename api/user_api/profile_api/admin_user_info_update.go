package profile_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/utils/maps"

	"github.com/gin-gonic/gin"
)

type AdminUserInfoUpdateRequest struct {
	UserID   uint           `json:"user_id" binding:"required"`
	Username *string        `json:"username"`
	Nickname *string        `json:"nickname"`
	Avatar   *string        `json:"avatar"`
	Abstract *string        `json:"abstract"`
	Role     *enum.RoleType `json:"role"`
}

func (ProfileApi) AdminUserInfoUpdateView(c *gin.Context) {
	cr := middleware.GetBindJson[AdminUserInfoUpdateRequest](c)

	userMap, err := maps.FieldsStructToMap(&cr, &models.UserModel{})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var userModel models.UserModel
	if err = global.DB.Take(&userModel, cr.UserID).Error; err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	if err = global.DB.Model(&userModel).Updates(userMap).Error; err != nil {
		res.FailWithMsg("用户信息更新失败", c)
		return
	}

	res.OkWithMsg("用户信息更新成功", c)
}
