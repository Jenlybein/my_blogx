package cron_service

import (
	"myblogx/global"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type CronService struct{}

func Cron() {
	// 创建调度器，指定上海时区（对应原代码的WithLocation）
	timezone, _ := time.LoadLocation("Asia/Shanghai")
	s, err := gocron.NewScheduler(
		gocron.WithLocation(timezone),
	)
	if err != nil {
		global.Logger.Errorf("创建gocron调度器失败: %v", err)
	}

	// 添加定时任务：每天2点执行SyncArticle
	_, err = s.NewJob(
		// 方式1：链式API（推荐，无需记crontab表达式）
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(2, 0, 0))),
		// gocron.DurationJob(2*time.Second),

		// 方式2：兼容原crontab表达式（如果想保留原有语法）
		// gocron.CronJob("0 0 2 * * *", false),

		// 指定要执行的任务函数
		gocron.NewTask(SyncArticle),
	)
	if err != nil {
		global.Logger.Errorf("添加同步文章任务失败: %v", err)
	}

	s.Start()
}
