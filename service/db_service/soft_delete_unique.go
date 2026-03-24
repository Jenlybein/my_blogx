package db_service

import (
	"errors"

	"gorm.io/gorm"
)

type UniqueWriteOptions struct {
	Model              any
	CreateValue        any
	Match              map[string]any
	RestoreAssignments map[string]any
}

// RestoreOrCreateUnique 按“先恢复软删，再尝试创建”的顺序执行。
// 返回 true 表示本次请求真正完成了“恢复”或“创建”，返回 false 表示活记录已存在。
func RestoreOrCreateUnique(tx *gorm.DB, opts UniqueWriteOptions) (bool, error) {
	if tx == nil {
		return false, errors.New("tx 不能为空")
	}

	// 先尝试恢复被软删除的记录
	restoreResult := tx.Unscoped().
		Model(opts.Model).
		Where(opts.Match).
		Where("deleted_at IS NOT NULL").
		Updates(opts.RestoreAssignments)
	if restoreResult.Error != nil {
		return false, restoreResult.Error
	}
	if restoreResult.RowsAffected > 0 {
		return true, nil
	}

	// 创建分支直接交给数据库唯一约束裁决。
	// 如果此时活记录已存在，或并发下别人刚创建成功，
	// 会返回 gorm.ErrDuplicatedKey，这里统一收敛为 false。
	createResult := tx.Create(opts.CreateValue)
	if createResult.Error != nil {
		if errors.Is(createResult.Error, gorm.ErrDuplicatedKey) {
			return false, nil
		}
		return false, createResult.Error
	}
	return true, nil
}
