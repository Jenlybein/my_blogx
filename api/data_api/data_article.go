package data_api

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models/enum"
	"time"

	"github.com/gin-gonic/gin"
)

func (DataApi) ArticleYearDataView(c *gin.Context) {
	var resp ArticleYearDataResponse

	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	rangeStart := currentMonthStart.AddDate(0, -11, 0)
	nextMonthStart := currentMonthStart.AddDate(0, 1, 0)

	var statList []DateCountItem
	err := global.DB.
		Table("article_models").
		Select("DATE_FORMAT(created_at, '%Y-%m') AS date, COUNT(*) AS count").
		Where("status = ? AND created_at >= ? AND created_at < ?", enum.ArticleStatusPublished, rangeStart, nextMonthStart).
		Group("DATE_FORMAT(created_at, '%Y-%m')").
		Order("DATE_FORMAT(created_at, '%Y-%m') ASC").
		Scan(&statList).Error
	if err != nil {
		global.Logger.Errorf("获取年度文章数据失败: %v", err)
		res.FailWithMsg("获取年度文章数据失败", c)
		return
	}

	dateMap := make(map[string]int, len(statList))
	for _, item := range statList {
		dateMap[item.Date] = item.Count
	}

	resp.DateCountList = make([]DateCountItem, 0, 12)
	for i := 0; i < 12; i++ {
		date := rangeStart.AddDate(0, i, 0).Format("2006-01")
		resp.DateCountList = append(resp.DateCountList, DateCountItem{
			Date:  date,
			Count: dateMap[date],
		})
	}

	res.OkWithData(resp, c)
}
