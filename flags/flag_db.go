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
		&models.UserStatModel{},
		&models.UserSessionModel{},
		&models.UserViewDailyModel{},
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
		&models.ImageRefModel{},
		&models.RuntimeSiteConfigModel{},
		&models.CommentModel{},
		&models.BannerModel{},
		&models.GlobalNotifModel{},
		&models.CommentDiggModel{},
		&models.ArticleMessageModel{},
		&models.UserGlobalNotifModel{},
		&models.UserFollowModel{},
		&models.ChatSessionModel{},
		&models.ChatMsgModel{},
		&models.ChatMsgUserStateModel{},
	)
	if err != nil {
		global.Logger.Error("数据库迁移失败", err)
		return
	}
	// if db.Migrator().HasTable("image_upload_task_models") {
	// 	if err := db.Migrator().DropTable("image_upload_task_models"); err != nil {
	// 		global.Logger.Errorf("删除旧图片上传任务表失败: %v", err)
	// 	}
	// }
	global.Logger.Info("数据库迁移成功")
}
