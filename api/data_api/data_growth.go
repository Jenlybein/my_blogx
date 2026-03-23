package data_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_site"
	"time"

	"github.com/gin-gonic/gin"
)

func (DataApi) GrowthDataView(c *gin.Context) {
	cr := middleware.GetBindQuery[GrowthDataRequest](c)
	var resp GrowthDataResponse

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rangeStart := todayStart.AddDate(0, 0, -6)
	tomorrowStart := todayStart.AddDate(0, 0, 1)

	var statList []DateCountItem
	var err error
	switch cr.Type {
	case 1:
		flowList := redis_site.GetRecentFlow(7)
		resp.DateCountList = make([]DateCountItem, 0, len(flowList))
		for _, item := range flowList {
			resp.DateCountList = append(resp.DateCountList, DateCountItem{
				Date:  item.Date,
				Count: item.Count,
			})
		}
	case 2:
		err = global.DB.
			Table("article_models").
			Select("DATE(created_at) AS date, COUNT(*) AS count").
			Where("status = ? AND created_at >= ? AND created_at < ?", enum.ArticleStatusPublished, rangeStart, tomorrowStart).
			Group("DATE(created_at)").
			Order("DATE(created_at) ASC").
			Scan(&statList).Error
	case 3:
		err = global.DB.
			Model(&models.UserModel{}).
			Select("DATE(created_at) AS date, COUNT(*) AS count").
			Where("created_at >= ? AND created_at < ?", rangeStart, tomorrowStart).
			Group("DATE(created_at)").
			Order("DATE(created_at) ASC").
			Scan(&statList).Error
	}
	if err != nil {
		global.Logger.Errorf("获取增长数据失败 type=%d: %v", cr.Type, err)
		res.FailWithMsg("获取增长数据失败", c)
		return
	}

	if cr.Type == 2 || cr.Type == 3 {
		dateMap := make(map[string]int, len(statList))
		for _, item := range statList {
			dateMap[item.Date] = item.Count
		}

		resp.DateCountList = make([]DateCountItem, 0, 7)
		for i := 0; i < 7; i++ {
			date := rangeStart.AddDate(0, 0, i).Format("2006-01-02")
			resp.DateCountList = append(resp.DateCountList, DateCountItem{
				Date:  date,
				Count: dateMap[date],
			})
		}
	}

	todayCount := resp.DateCountList[len(resp.DateCountList)-1].Count
	yesterdayCount := resp.DateCountList[len(resp.DateCountList)-2].Count
	resp.GrowthNum = todayCount - yesterdayCount
	if yesterdayCount == 0 {
		if todayCount > 0 {
			resp.GrowthRate = 100
		}
	} else {
		resp.GrowthRate = int(float64(resp.GrowthNum) / float64(yesterdayCount) * 100)
	}

	res.OkWithData(resp, c)
}
