package models

type UserFollowModel struct {
	Model
	FollowedUserID    uint      `gorm:"uniqueIndex:uk_user_follow,priority:1" json:"followed_user_id"`
	FansUserID        uint      `gorm:"uniqueIndex:uk_user_follow,priority:2" json:"fans_user_id"`
	FollowedUserModel UserModel `gorm:"foreignKey:FollowedUserID;references:ID"`
	FansUserModel     UserModel `gorm:"foreignKey:FansUserID;references:ID"`
}
