package models

import "time"

type UserGlobalNotifModel struct {
	Model
	MsgID  uint       `json:"msg_id"`
	UserID uint       `json:"user_id"`
	IsRead bool       `json:"is_read"`
	ReadAt *time.Time `json:"read_at"`
}
