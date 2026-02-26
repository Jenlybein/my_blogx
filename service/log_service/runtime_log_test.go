package log_service_test

import (
	"myblogx/service/log_service"
	"testing"
)

func TestRuntimeDateTypeGetSqlTime(t *testing.T) {
	cases := map[log_service.RuntimeDateType]string{
		log_service.RuntimeDateHour:  "interval 1 HOUR",
		log_service.RuntimeDateDay:   "interval 1 DAY",
		log_service.RuntimeDateWeek:  "interval 1 WEEK",
		log_service.RuntimeDateMonth: "interval 1 MONTH",
	}

	for in, want := range cases {
		if got := in.GetSqlTime(); got != want {
			t.Fatalf("GetSqlTime 错误: in=%v got=%s want=%s", in, got, want)
		}
	}

	if got := log_service.RuntimeDateType(99).GetSqlTime(); got != "interval 1 DAY" {
		t.Fatalf("默认值错误: %s", got)
	}
}

func TestNewRuntimeLog(t *testing.T) {
	rl := log_service.NewRuntimeLog("sync_service", log_service.RuntimeDateWeek)
	if rl == nil {
		t.Fatal("NewRuntimeLog 不应返回 nil")
	}
	if rl.RuntimeDateType != log_service.RuntimeDateWeek {
		t.Fatalf("RuntimeDateType 异常: %v", rl.RuntimeDateType)
	}
}
