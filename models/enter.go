// 模型模块基础定义

package models

import (
	"time"
)

// 基础模型
type Model struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// ID请求参数
type IDRequest struct {
	ID uint `json:"id" form:"id" uri:"id"`
}

type RemoveRequest struct {
	IDList []uint `json:"id_list" binding:"required"`
}

type OptionsResponse[T any] struct {
	Label string `json:"label"`
	Value T      `json:"value"`
}
