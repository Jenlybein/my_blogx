package follow_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/service/follow_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

// TODO：增加用户名搜索
// FollowListView 获取关注列表
func (f *FollowApi) FollowListView(c *gin.Context) {
	cr := middleware.GetBindQuery[FollowListRequest](c)

	claims := jwts.GetClaimsByGin(c)

	// 查询目标用户隐私设置，判断是否公开关注列表
	if cr.UserID != claims.UserID {
		if cr.UserID != 0 {
			var user models.UserConfModel
			if err := global.DB.Take(&user, "user_id = ?", cr.UserID).Error; err != nil {
				res.FailWithMsg("用户配置信息不存在", c)
				return
			}
			if !user.FollowVisibility {
				res.FailWithMsg("关注列表不公开", c)
				return
			}
		} else {
			cr.UserID = claims.UserID
		}
	}

	_list, count, err := common.ListQuery(models.UserFollowModel{
		FollowedUserID: cr.FollowedUserID,
		FansUserID:     cr.UserID,
	}, common.Options{
		PageInfo:      cr.PageInfo,
		ExactPreloads: map[string][]string{"FollowedUserModel": {"id", "avatar", "nickname", "abstract", "created_at"}},
	})

	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 计算用户关系
	userIDs := make([]ctype.ID, 0, len(_list))
	for _, item := range _list {
		userIDs = append(userIDs, item.FollowedUserID)
	}
	relationMap := follow_service.CalUserRelationshipBatch(claims.UserID, userIDs)

	// 格式化
	var list = make([]FollowListResponse, 0)
	for _, item := range _list {
		list = append(list, FollowListResponse{
			FollowedUserID:   item.FollowedUserID,
			FollowedNickname: item.FollowedUserModel.Nickname,
			FollowedAvatar:   item.FollowedUserModel.Avatar,
			FollowedAbstract: item.FollowedUserModel.Abstract,
			FollowTime:       item.CreatedAt,
			Relation:         int8(relationMap[item.FollowedUserID]),
		})
	}
	res.OkWithList(list, count, c)
}
