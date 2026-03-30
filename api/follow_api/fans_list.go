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
// FansListView 获取粉丝列表
func (f *FollowApi) FansListView(c *gin.Context) {
	cr := middleware.GetBindQuery[FansListRequest](c)

	claims := jwts.GetClaimsByGin(c)

	// 查询目标用户隐私设置，判断是否公开粉丝列表
	if cr.UserID != claims.UserID {
		if cr.UserID != 0 {
			var user models.UserConfModel
			if err := global.DB.Take(&user, "user_id = ?", cr.UserID).Error; err != nil {
				res.FailWithMsg("用户配置信息不存在", c)
				return
			}
			if !user.FansVisibility {
				res.FailWithMsg("粉丝列表不公开", c)
				return
			}
		} else {
			cr.UserID = claims.UserID
		}
	}

	_list, count, err := common.ListQuery(models.UserFollowModel{
		FollowedUserID: cr.UserID,
		FansUserID:     cr.FansUserID,
	}, common.Options{
		PageInfo:      cr.PageInfo,
		ExactPreloads: map[string][]string{"FansUserModel": {"id", "avatar", "nickname", "abstract", "created_at"}},
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var list = make([]FansListResponse, 0, len(_list))
	userIDs := make([]ctype.ID, 0, len(_list))
	for _, item := range _list {
		userIDs = append(userIDs, item.FansUserID)
	}
	relationMap := follow_service.CalUserRelationshipBatch(claims.UserID, userIDs)
	for _, item := range _list {
		list = append(list, FansListResponse{
			FansUserID:   item.FansUserID,
			FansNickname: item.FansUserModel.Nickname,
			FansAvatar:   item.FansUserModel.Avatar,
			FansAbstract: item.FansUserModel.Abstract,
			FollowTime:   item.CreatedAt,
			Relation:     int8(relationMap[item.FansUserID]),
		})
	}
	res.OkWithList(list, count, c)
}
