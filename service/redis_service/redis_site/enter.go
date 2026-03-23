package redis_site

import (
	"context"
	"myblogx/global"
	"strconv"
	"time"
)

const (
	siteFlowTotalKey = "blog_site_flow:total"
	siteFlowDailyKey = "blog_site_flow:daily"
	maxDailyFlowDays = 30
)

type DateCountItem struct {
	Date  string
	Count int
}

// SetFlow 写入站点流量
func SetFlow() {
	if global.Redis == nil {
		return
	}

	ctx := context.Background()
	now := time.Now()
	today := now.Format("2006-01-02")

	pipe := global.Redis.TxPipeline()
	pipe.IncrBy(ctx, siteFlowTotalKey, 1)
	pipe.HIncrBy(ctx, siteFlowDailyKey, today, 1)

	if _, err := pipe.Exec(ctx); err != nil {
		global.Logger.Errorf("站点流量写入失败: %v", err)
		return
	}
	pruneExpiredDailyFlow(now)
}

// GetFlow 读取站点流量
func GetFlow() int {
	if global.Redis == nil {
		return 0
	}
	num, _ := global.Redis.Get(context.Background(), siteFlowTotalKey).Int()
	return num
}

// GetRecentFlow 获取最近 days 天的站点流量
func GetRecentFlow(days int) []DateCountItem {
	if global.Redis == nil || days <= 0 {
		return nil
	}

	now := time.Now()
	pruneExpiredDailyFlow(now)

	values, err := global.Redis.HGetAll(context.Background(), siteFlowDailyKey).Result()
	if err != nil {
		global.Logger.Errorf("读取站点流量失败: %v", err)
		return nil
	}

	dateCountList := make([]DateCountItem, 0, days)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(days - 1))
	for i := 0; i < days; i++ {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		count, _ := strconv.Atoi(values[date])
		dateCountList = append(dateCountList, DateCountItem{
			Date:  date,
			Count: count,
		})
	}
	return dateCountList
}

// pruneExpiredDailyFlow 清理过期的站点流量
func pruneExpiredDailyFlow(now time.Time) {
	if global.Redis == nil {
		return
	}

	keys, err := global.Redis.HKeys(context.Background(), siteFlowDailyKey).Result()
	if err != nil {
		global.Logger.Errorf("读取站点流量日期失败: %v", err)
		return
	}

	expiredDates := buildExpiredDates(now, keys)
	if len(expiredDates) == 0 {
		return
	}

	fields := make([]string, 0, len(expiredDates))
	for _, date := range expiredDates {
		fields = append(fields, date)
	}
	if err := global.Redis.HDel(context.Background(), siteFlowDailyKey, fields...).Err(); err != nil {
		global.Logger.Errorf("清理过期站点流量失败: %v", err)
	}
}

func buildExpiredDates(now time.Time, keys []string) []string {
	expiredDates := make([]string, 0)
	cutoff := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(maxDailyFlowDays - 1))
	for _, key := range keys {
		date, err := time.ParseInLocation("2006-01-02", key, now.Location())
		if err != nil {
			expiredDates = append(expiredDates, key)
			continue
		}
		if date.Before(cutoff) {
			expiredDates = append(expiredDates, key)
		}
	}
	return expiredDates
}
