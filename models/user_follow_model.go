package models

import "myblogx/models/ctype"

type UserFollowModel struct {
	Model
	FollowedUserID    ctype.ID  `gorm:"uniqueIndex:uk_user_follow,priority:1" json:"followed_user_id"`
	FansUserID        ctype.ID  `gorm:"uniqueIndex:uk_user_follow,priority:2" json:"fans_user_id"`
	FollowedUserModel UserModel `gorm:"foreignKey:FollowedUserID;references:ID"`
	FansUserModel     UserModel `gorm:"foreignKey:FansUserID;references:ID"`
}
