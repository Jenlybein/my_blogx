package flags

import (
	"myblogx/global"
	"myblogx/models"

	"gorm.io/gorm"
)

func FlagDB(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ArticleModel{},
		&models.TagModel{},
		&models.ArticleTagModel{},
		&models.ArticleDiggModel{},
		&models.CategoryModel{},
		&models.FavoriteModel{},
		&models.UserArticleFavorModel{},
		&models.UserArticleViewHistoryModel{},
		&models.UserTopArticleModel{},
		&models.ImageModel{},
		&models.CommentModel{},
		&models.LogModel{},
		&models.BannerModel{},
		&models.UserLoginModel{},
		&models.GlobalNotificationModel{},
		&models.CommentDiggModel{},
		&models.ArticleMessageModel{},
	)
	if err != nil {
		global.Logger.Error("数据库迁移失败", err)
		return
	}
	global.Logger.Info("数据库迁移成功")
}
