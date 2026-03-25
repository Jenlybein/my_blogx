package profile_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/user_service"
	"myblogx/utils/maps"

	"github.com/gin-gonic/gin"
)

type AdminUserInfoUpdateRequest struct {
	UserID   ctype.ID         `json:"user_id" binding:"required"`
	Username *string          `json:"username"`
	Nickname *string          `json:"nickname"`
	Avatar   *string          `json:"avatar"`
	Abstract *string          `json:"abstract"`
	Role     *enum.RoleType   `json:"role"`
	Status   *enum.UserStatus `json:"status"`
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

	if (cr.Role != nil && *cr.Role != userModel.Role) || (cr.Status != nil && *cr.Status != userModel.Status) {
		if err = user_service.InvalidateUserAuthState(&userModel); err != nil {
			res.FailWithMsg("用户信息更新成功，但会话失效处理失败", c)
			return
		}
	}

	res.OkWithMsg("用户信息更新成功", c)
}
