package follow_service

import (
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum/relationship_enum"
)

// CalUserRelationship 计算用户关系
func CalUserRelationship(A, B ctype.ID) relationship_enum.Relation {
	if A == 0 || B == 0 {
		return relationship_enum.RelationStranger
	}

	return CalUserRelationshipBatch(A, []ctype.ID{B})[B]
}

// 批量计算用户关系
func CalUserRelationshipBatch(user ctype.ID, userList []ctype.ID) map[ctype.ID]relationship_enum.Relation {
	relationMap := make(map[ctype.ID]relationship_enum.Relation, len(userList))
	if len(userList) == 0 {
		return relationMap
	}

	for _, other := range userList {
		relationMap[other] = relationship_enum.RelationStranger
	}

	var rows []models.UserFollowModel
	if err := global.DB.
		Where("(followed_user_id = ? AND fans_user_id IN ?) OR (followed_user_id IN ? AND fans_user_id = ?)",
			user, userList, userList, user).
		Find(&rows).Error; err != nil {
		return relationMap
	}

	const (
		iFollow  uint8 = 1 // 当前 user 关注了对方
		heFollow uint8 = 2 // 对方关注了当前 user
	)
	state := make(map[ctype.ID]uint8, len(userList))

	for _, row := range rows {
		switch {
		case row.FansUserID == user:
			state[row.FollowedUserID] |= iFollow
		case row.FollowedUserID == user:
			state[row.FansUserID] |= heFollow
		}
	}

	for other, s := range state {
		switch s {
		case iFollow:
			relationMap[other] = relationship_enum.RelationFollowed
		case heFollow:
			relationMap[other] = relationship_enum.RelationFans
		case iFollow | heFollow:
			relationMap[other] = relationship_enum.RelationFriend
		}
	}

	return relationMap
}
