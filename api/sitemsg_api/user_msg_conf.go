package sitemsg_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/jwts"
	"myblogx/utils/maps"

	"github.com/gin-gonic/gin"
)

func (a *SitemsgApi) UserMsgConfView(c *gin.Context) {
	claims := jwts.GetClaimsByGin(c)

	var userConfModel models.UserConfModel
	if err := global.DB.Take(&userConfModel, claims.UserID).Error; err != nil {
		res.FailWithMsg("用户配置信息不存在", c)
		return
	}

	msgConf := UserMsgConfResponseAndRequest{
		DiggNoticeEnabled:        userConfModel.DiggNoticeEnabled,
		CommentNoticeEnabled:     userConfModel.CommentNoticeEnabled,
		FavorNoticeEnabled:       userConfModel.FavorNoticeEnabled,
		PrivateChatNoticeEnabled: userConfModel.PrivateChatNoticeEnabled,
	}

	res.OkWithData(msgConf, c)
}

func (a *SitemsgApi) UserMsgConfUpdateView(c *gin.Context) {
	cr := middleware.GetBindJson[UserMsgConfResponseAndRequest](c)

	claims := jwts.GetClaimsByGin(c)

	confMap, err := maps.FieldsStructToMap(&cr, &models.UserConfModel{})
	if err != nil {
		res.FailWithError(err, c)
		return
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

	res.OkWithMsg("用户配置信息更新成功", c)
}
