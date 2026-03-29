package user_man_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/service/log_service"
	"time"

	"github.com/gin-gonic/gin"
)

func (a *UserManApi) UserListView(c *gin.Context) {
	cr := middleware.GetBindQuery[UserListRequest](c)

	_list, count, _ := common.ListQuery(models.UserModel{}, common.Options{
		Likes:    []string{"nickname", "username"},
		PageInfo: cr.PageInfo,
	})

	idList := make([]ctype.ID, 0, len(_list))
	for _, item := range _list {
		idList = append(idList, item.ID)
	}
	latestLoginMap, err := log_service.LoadLatestLoginMap(idList)
	if err != nil {
		global.Logger.Errorf("加载用户最后登录信息失败: %v", err)
	}

	var list = make([]UserListResponse, 0)
	for _, item := range _list {
		data := UserListResponse{
			ID:        item.ID,
			Nickname:  item.Nickname,
			Avatar:    item.Avatar,
			Username:  item.Username,
			CreatedAt: item.CreatedAt,
		}
		if lastLogin, ok := latestLoginMap[item.ID]; ok {
			data.IP = lastLogin.IP
			data.Addr = lastLogin.Addr
			if parsedAt, parseErr := time.ParseInLocation("2006-01-02 15:04:05.000", lastLogin.TS, time.Local); parseErr == nil {
				data.LastLoginAt = parsedAt
			}
		}
		list = append(list, data)
	}

	res.OkWithList(list, count, c)
}
