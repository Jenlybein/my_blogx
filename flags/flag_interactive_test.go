package flags_test

import (
	"myblogx/conf"
	"myblogx/flags"
	"myblogx/global"
	"myblogx/models"
	"myblogx/test/testutil"
	"os"
	"testing"
)

func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()
	old := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("创建管道失败: %v", err)
	}
	if _, err = w.WriteString(input); err != nil {
		t.Fatalf("写入 stdin 数据失败: %v", err)
	}
	_ = w.Close()
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = old
		_ = r.Close()
	})
	fn()
}

func TestFlagESIndexNoOp(t *testing.T) {
	testutil.InitGlobals()
	global.Config = &conf.Config{
		ES: conf.ES{Index: "article_idx"},
	}

	withStdin(t, "3\n3\n", func() {
		flags.FlagESIndex()
	})
}

func TestFlagUserCreateInvalidRoleAndExistsUser(t *testing.T) {
	db := testutil.SetupSQLite(t, &models.UserModel{}, &models.UserConfModel{})

	t.Run("非法角色直接返回", func(t *testing.T) {
		withStdin(t, "0\n", func() {
			u := flags.FlagUser{}
			u.Create(db)
		})
		var cnt int64
		_ = db.Model(&models.UserModel{}).Count(&cnt).Error
		if cnt != 0 {
			t.Fatalf("非法角色不应创建用户, cnt=%d", cnt)
		}
	})

	t.Run("用户名已存在直接返回", func(t *testing.T) {
		exists := models.UserModel{
			Username: "exists_u",
			Password: "x",
		}
		if err := db.Create(&exists).Error; err != nil {
			t.Fatalf("创建已有用户失败: %v", err)
		}

		withStdin(t, "1\nexists_u\n", func() {
			u := flags.FlagUser{}
			u.Create(db)
		})

		var cnt int64
		_ = db.Model(&models.UserModel{}).Where("username = ?", "exists_u").Count(&cnt).Error
		if cnt != 1 {
			t.Fatalf("已存在用户分支不应重复创建, cnt=%d", cnt)
		}
	})
}
