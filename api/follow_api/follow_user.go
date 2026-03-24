package follow_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	dbservice "myblogx/service/db_service"
	"myblogx/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
)

// 当前登录用户关注其他用户
func (FollowApi) FollowUserView(c *gin.Context) {
	cr := middleware.GetBindUri[models.IDRequest](c)

	claims := jwts.GetClaimsByGin(c)

	if cr.ID == claims.UserID {
		res.FailWithMsg("不能关注自己", c)
		return
	}

	// TODO：考虑每天关注量上限和取关量上限

	// 这里不能再依赖“先查一次”来判断是否成功，而是要看本次写入是否真的生效。
	createdOrRestored, err := dbservice.RestoreOrCreateUnique(global.DB, dbservice.UniqueWriteOptions{
		Model: &models.UserFollowModel{},
		CreateValue: &models.UserFollowModel{
			FollowedUserID: cr.ID,
			FansUserID:     claims.UserID,
		},
		Match: map[string]any{
			"followed_user_id": cr.ID,
			"fans_user_id":     claims.UserID,
		},
		RestoreAssignments: map[string]any{
			"deleted_at": nil,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		res.FailWithMsg("关注失败", c)
		return
	}
	if !createdOrRestored {
		res.FailWithMsg("请勿重复关注", c)
		return
	}
	res.OkWithMsg("关注成功", c)
}

// 当前登录用户取消关注其他用户
func (FollowApi) UnfollowUserView(c *gin.Context) {
	cr := middleware.GetBindUri[models.IDRequest](c)

	claims := jwts.GetClaimsByGin(c)

	if cr.ID == claims.UserID {
		res.FailWithMsg("不能取消关注自己", c)
		return
	}

	// 取消关注必须看本次 Delete 是否真正命中了活记录，不能只看删除前查到了什么。
	deleteResult := global.DB.Where(map[string]any{
		"followed_user_id": cr.ID,
		"fans_user_id":     claims.UserID,
	}).Delete(&models.UserFollowModel{})
	if deleteResult.Error != nil {
		res.FailWithMsg("取消关注失败", c)
		return
	}
	if deleteResult.RowsAffected == 0 {
		res.FailWithMsg("尚未关注该用户", c)
		return
	}
	res.OkWithMsg("取消关注成功", c)
}
