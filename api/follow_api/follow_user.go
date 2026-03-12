package follow_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	var follow models.UserFollowModel
	if err := global.DB.Take(&follow, "followed_user_id = ? and fans_user_id = ?", cr.ID, claims.UserID).Error; err == nil {
		res.FailWithMsg("请勿重复关注", c)
		return
	}

	if err := global.DB.Create(&models.UserFollowModel{
		FollowedUserID: cr.ID,
		FansUserID:     claims.UserID,
	}).Error; err != nil {
		res.FailWithMsg("关注失败", c)
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

	var follow models.UserFollowModel
	if err := global.DB.Take(&follow, "followed_user_id = ? and fans_user_id = ?", cr.ID, claims.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			res.FailWithMsg("尚未关注该用户", c)
			return
		}
		res.FailWithMsg("查询关注关系失败", c)
		return
	}

	if err := global.DB.Unscoped().Delete(&follow).Error; err != nil {
		res.FailWithMsg("取消关注失败", c)
		return
	}
	res.OkWithMsg("取消关注成功", c)
}
