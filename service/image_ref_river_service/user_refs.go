package image_ref_river_service

import (
	"strings"

	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum/image_ref_enum"

	"gorm.io/gorm"
)

func RebuildUserRefs(tx *gorm.DB, user *models.UserModel) error {
	if tx == nil {
		tx = global.DB
	}
	candidates := make([]refCandidate, 0, 1)
	if avatar := strings.TrimSpace(user.Avatar); avatar != "" {
		candidates = append(candidates, refCandidate{
			Field:    image_ref_enum.RefFieldUserAvatar,
			Position: 0,
			URL:      avatar,
		})
	}
	return replaceOwnerRefs(tx, image_ref_enum.RefTypeUser, user.ID, candidates)
}

func RebuildUserRefsByRow(snapshot rowSnapshot) error {
	userID, err := snapshot.ID()
	if err != nil {
		return err
	}
	if snapshot.IsDeleted() {
		return DeleteOwnerRefs(global.DB, image_ref_enum.RefTypeUser, userID)
	}
	avatar, err := snapshot.RequireString("avatar")
	if err != nil {
		return err
	}
	return RebuildUserRefs(global.DB, &models.UserModel{
		Model:  models.Model{ID: userID},
		Avatar: avatar,
	})
}
