package banner_api

import (
	"fmt"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

type BannerApi struct{}

type BannerCreateRequest struct {
	Cover string `json:"cover" binding:"required"`
	Href  string `json:"href"`
	Show  bool   `json:"show"`
}

func (BannerApi) BannerCreateView(c *gin.Context) {
	cr := middleware.GetBindJson[BannerCreateRequest](c)

	// 创建轮播图
	if err := global.DB.Create(&models.BannerModel{
		Cover: cr.Cover,
		Href:  cr.Href,
		Show:  cr.Show,
	}).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	res.OkWithMsg("创建轮播图成功", c)
}

type BannerListRequest struct {
	common.PageInfo
	Show bool `form:"show"`
}

func (BannerApi) BannerListView(c *gin.Context) {
	cr := middleware.GetBindQuery[BannerListRequest](c)

	list, count, err := common.ListQuery(models.BannerModel{
		Show: cr.Show,
	}, common.Options{
		PageInfo: cr.PageInfo,
	})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	res.OkWithList(list, count, c)
}

func (BannerApi) BannerRemoveView(c *gin.Context) {
	cr := middleware.GetBindJson[models.IDListRequest](c)

	var list []models.BannerModel
	if err := global.DB.Find(&list, "id IN ?", cr.IDList).Error; err != nil {
		res.FailWithError(err, c)
		return
	}
	if len(list) > 0 {
		if err := global.DB.Delete(&list).Error; err != nil {
			res.FailWithError(err, c)
			return
		}
	}

	res.OkWithMsg(fmt.Sprintf("请求删除轮播图%d个, 成功%d条", len(cr.IDList), len(list)), c)
}

func (BannerApi) BannerUpdateView(c *gin.Context) {
	id := middleware.GetBindUri[models.IDRequest](c)

	cr := middleware.GetBindJson[BannerCreateRequest](c)

	var model models.BannerModel
	if err := global.DB.Take(&model, id.ID).Error; err != nil {
		res.FailWithMsg("轮播图不存在", c)
		return
	}

	if err := global.DB.Model(&model).Updates(models.BannerModel{
		Cover: cr.Cover,
		Href:  cr.Href,
		Show:  cr.Show,
	}).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	res.OkWithMsg("更新轮播图成功", c)
}
