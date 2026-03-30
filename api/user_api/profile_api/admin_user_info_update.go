package profile_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/log_service"
	"myblogx/service/user_service"
	"myblogx/utils/maps"
	"strconv"

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
	log_service.EmitActionAuditFromGin(c, log_service.GinAuditInput{
		ActionName: "admin_user_update",
		TargetType: "user",
		TargetID:   strconv.FormatUint(uint64(cr.UserID), 10),
		Success:    true,
		Message:    "管理员更新用户信息成功",
		RequestBody: map[string]any{
			"user_id":  cr.UserID,
			"username": cr.Username,
			"nickname": cr.Nickname,
			"avatar":   cr.Avatar,
			"abstract": cr.Abstract,
			"role":     cr.Role,
			"status":   cr.Status,
		},
		UseRawRequestBody: true,
		UseRawRequestHead: true,
	})
}
