package db_service

import (
	"testing"
	"time"

	"myblogx/models"
	"myblogx/test/testutil"
)

type softDeleteUniqueTestModel struct {
	models.Model
	UserID uint   `gorm:"uniqueIndex:uk_soft_delete_unique_test,priority:1"`
	Title  string `gorm:"size:64;uniqueIndex:uk_soft_delete_unique_test,priority:2"`
	Remark string
}

func TestRestoreOrCreateUnique(t *testing.T) {
	db := testutil.SetupSQLite(t, &softDeleteUniqueTestModel{})

	t.Run("创建新记录", func(t *testing.T) {
		ok, err := RestoreOrCreateUnique(db, UniqueWriteOptions{
			Model: &softDeleteUniqueTestModel{},
			CreateValue: &softDeleteUniqueTestModel{
				UserID: 1,
				Title:  "created",
				Remark: "v1",
			},
			Match: map[string]any{
				"user_id": 1,
				"title":   "created",
			},
			RestoreAssignments: map[string]any{
				"deleted_at": nil,
				"updated_at": time.Now(),
				"remark":     "v1",
			},
		})
		if err != nil {
			t.Fatalf("创建新记录失败: %v", err)
		}
		if !ok {
			t.Fatal("创建结果错误: 预期本次写入成功")
		}
	})

	t.Run("恢复软删记录", func(t *testing.T) {
		row := softDeleteUniqueTestModel{
			UserID: 2,
			Title:  "restored",
			Remark: "old",
		}
		if err := db.Create(&row).Error; err != nil {
			t.Fatalf("准备软删记录失败: %v", err)
		}
		if err := db.Delete(&row).Error; err != nil {
			t.Fatalf("软删记录失败: %v", err)
		}

		now := time.Now()
		ok, err := RestoreOrCreateUnique(db, UniqueWriteOptions{
			Model: &softDeleteUniqueTestModel{},
			CreateValue: &softDeleteUniqueTestModel{
				UserID: 2,
				Title:  "restored",
				Remark: "new",
			},
			Match: map[string]any{
				"user_id": 2,
				"title":   "restored",
			},
			RestoreAssignments: map[string]any{
				"deleted_at": nil,
				"updated_at": now,
				"remark":     "new",
			},
		})
		if err != nil {
			t.Fatalf("恢复软删记录失败: %v", err)
		}
		if !ok {
			t.Fatal("恢复结果错误: 预期本次写入成功")
		}

		var restored softDeleteUniqueTestModel
		if err := db.Unscoped().Take(&restored, row.ID).Error; err != nil {
			t.Fatalf("回查恢复记录失败: %v", err)
		}
		if restored.DeletedAt.Valid {
			t.Fatal("恢复后 deleted_at 应为空")
		}
		if restored.Remark != "new" {
			t.Fatalf("恢复后 remark 未更新: got=%s", restored.Remark)
		}
	})

	t.Run("活记录已存在时返回 existing", func(t *testing.T) {
		row := softDeleteUniqueTestModel{
			UserID: 3,
			Title:  "existing",
			Remark: "keep",
		}
		if err := db.Create(&row).Error; err != nil {
			t.Fatalf("准备活记录失败: %v", err)
		}

		ok, err := RestoreOrCreateUnique(db, UniqueWriteOptions{
			Model: &softDeleteUniqueTestModel{},
			CreateValue: &softDeleteUniqueTestModel{
				UserID: 3,
				Title:  "existing",
				Remark: "new",
			},
			Match: map[string]any{
				"user_id": 3,
				"title":   "existing",
			},
			RestoreAssignments: map[string]any{
				"deleted_at": nil,
				"updated_at": time.Now(),
				"remark":     "new",
			},
		})
		if err != nil {
			t.Fatalf("existing 分支失败: %v", err)
		}
		if ok {
			t.Fatal("existing 结果错误: 活记录已存在时不应返回成功")
		}
	})
}
