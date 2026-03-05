package flags

import (
	"myblogx/global"
	"myblogx/models"

	"gorm.io/gorm"
)

func FlagDB(db *gorm.DB) {
	// 自动建表
	err := db.AutoMigrate(
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ArticleModel{},
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
	)
	if err != nil {
		global.Logger.Error("数据库迁移失败: ", err)
		return
	}
	global.Logger.Info("数据库迁移成功")
}
