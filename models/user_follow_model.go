package models

import "time"

type UserFollowModel struct {
	ID                int64     `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	FollowedUserID    uint     `json:"followed_user_id"`
	FansUserID        uint     `json:"fans_user_id"`
	FollowedUserModel UserModel `gorm:"foreignKey:FollowedUserID;references:ID"`
	FansUserModel     UserModel `gorm:"foreignKey:FansUserID;references:ID"`
}
