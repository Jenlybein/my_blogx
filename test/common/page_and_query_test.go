package common_test

import (
	"myblogx/common"
	"myblogx/models"
	"myblogx/test/testutil"
	"testing"
)

func TestPageInfoHelpers(t *testing.T) {
	p := common.PageInfo{Page: -1, Limit: 1000}
	if p.GetPage() != 1 {
		t.Fatalf("GetPage 默认值错误: %d", p.GetPage())
	}
	if p.GetLimit() != 10 {
		t.Fatalf("GetLimit 默认值错误: %d", p.GetLimit())
	}
	if p.GetOffset() != 0 {
		t.Fatalf("GetOffset 错误: %d", p.GetOffset())
	}

	p = common.PageInfo{Page: 2, Limit: 5}
	if p.GetOffset() != 5 {
		t.Fatalf("GetOffset 计算错误: %d", p.GetOffset())
	}
}

func TestListQueryBasic(t *testing.T) {
	db := testutil.SetupSQLite(t, &models.BannerModel{})

	records := []models.BannerModel{
		{Show: true, Cover: "alpha-cover", Href: "/a"},
		{Show: true, Cover: "beta-cover", Href: "/b"},
		{Show: false, Cover: "gamma", Href: "/c"},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	list, count, err := common.ListQuery(
		models.BannerModel{},
		common.Options{
			PageInfo: common.PageInfo{Page: 1, Limit: 2, Key: "cover"},
			Likes:    []string{"cover"},
		},
	)
	if err != nil {
		t.Fatalf("ListQuery 查询失败: %v", err)
	}
	if count != 2 {
		t.Fatalf("count 错误: %d", count)
	}
	if len(list) != 2 {
		t.Fatalf("分页长度错误: %d", len(list))
	}
}

func TestListQueryInvalidOrder(t *testing.T) {
	_ = testutil.SetupSQLite(t, &models.BannerModel{})

	_, _, err := common.ListQuery(
		models.BannerModel{},
		common.Options{
			PageInfo: common.PageInfo{Order: "created_at desc"},
			OrderMap: map[string]bool{
				"id desc": true,
			},
		},
	)
	if err == nil {
		t.Fatal("非法排序字段应报错")
	}
}
