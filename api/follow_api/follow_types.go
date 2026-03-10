package follow_api

import (
	"myblogx/common"
	"time"
)

type FollowListRequest struct {
	common.PageInfo
	FollowedUserID uint `form:"followed_user_id"`
	UserID         uint `form:"user_id"`
}

type FollowListResponse struct {
	FollowedUserID   uint      `json:"followed_user_id"`
	FollowedNickname string    `json:"followed_nickname"`
	FollowedAvatar   string    `json:"followed_avatar"`
	FollowedAbstract string    `json:"followed_abstract"`
	FollowTime       time.Time `json:"follow_time"`
}

type FansListRequest struct {
	common.PageInfo
	FansUserID uint `form:"fans_user_id"`
	UserID     uint `form:"user_id"`
}

type FansListResponse struct {
	FansUserID   uint      `json:"fans_user_id"`
	FansNickname string    `json:"fans_nickname"`
	FansAvatar   string    `json:"fans_avatar"`
	FansAbstract string    `json:"fans_abstract"`
	FollowTime   time.Time `json:"follow_time"`
}
