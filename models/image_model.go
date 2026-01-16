// 图片模型

package models

import (
	"fmt"
	"myblogx/global"
	"os"

	"gorm.io/gorm"
)

// 图片表
type ImageModel struct {
	Model
	FileName string `gorm:"size:64" json:"file_name"`
	Path     string `gorm:"size:256" json:"path"`
	Size     int64  `gorm:"default:0" json:"size"` // 图片大小
	Hash     string `gorm:"size:32" json:"hash"`   // 图片哈希值
}

func (i ImageModel) WebPath() string {
	return fmt.Sprintf("/%s", i.Path)
}

// Gorm 固定命名的 Hook，调用 Delete 会自动调用该方法
func (i *ImageModel) BeforeDelete(tx *gorm.DB) error {
	if err := os.Remove(i.Path); err != nil {
		global.Logger.Warnf("删除图片文件失败: %v", err)
	}
	return nil
}
