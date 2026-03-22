package user_man_api

import (
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func (a *UserManApi) UserListView(c *gin.Context) {
	cr := middleware.GetBindQuery[UserListRequest](c)

	_list, count, _ := common.ListQuery(models.UserModel{}, common.Options{
		Likes:         []string{"nickname", "username"},
		ExactPreloads: map[string][]string{"LoginList": {"id", "ip", "addr", "created_at"}},
		PageInfo:      cr.PageInfo,
	})

	var list = make([]UserListResponse, 0)
	for _, item := range _list {
		data := UserListResponse{
			ID:        item.ID,
			Nickname:  item.Nickname,
			Avatar:    item.Avatar,
			Username:  item.Username,
			CreatedAt: item.CreatedAt,
		}
		if item.LoginList != nil {
			lastLogin := item.LoginList[len(item.LoginList)-1]
			data.IP = lastLogin.IP
			data.Addr = lastLogin.Addr
			data.LastLoginAt = lastLogin.CreatedAt
		}
		list = append(list, data)
	}

	res.OkWithList(list, count, c)
}
