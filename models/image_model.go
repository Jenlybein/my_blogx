// 图片模型

package models

// 图片表
type ImageModel struct {
	Model
	FileName string `gorm:"size:64" json:"file_name"`
	Path     string `gorm:"size:256" json:"path"`
	Size     int    `gorm:"default:0" json:"size"` // 图片大小
	Hash     string `gorm:"size:32" json:"hash"`   // 图片哈希值
}
