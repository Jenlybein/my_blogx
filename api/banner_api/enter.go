package banner_api

import (
	"fmt"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BannerApi struct{}

type BannerCreateRequest struct {
	Cover string `json:"cover" binding:"required"`
	Href  string `json:"href"`
	Show  bool   `json:"show"`
}

func (BannerApi) BannerCreateView(c *gin.Context) {
	var cr BannerCreateRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
	Show bool `json:"show"`
}

func (BannerApi) BannerListView(c *gin.Context) {
	var cr BannerListRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, c)
		return
	}

	list, count, _ := common.ListQuery(models.BannerModel{
		Show: cr.Show,
	}, common.Options{
		PageInfo: cr.PageInfo,
	})

	res.OkWithList(list, count, c)
}

func (BannerApi) BannerRemoveView(c *gin.Context) {
	var cr models.RemoveRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, c)
		return
	}

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
	var id models.IDRequest
	if err := c.ShouldBindUri(&id); err != nil {
		res.FailWithError(err, c)
		return
	}

	var cr BannerCreateRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		res.FailWithError(err, c)
		return
	}

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
