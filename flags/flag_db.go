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
		&models.UserSessionModel{},
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
		&models.CommentModel{},
		&models.LogModel{},
		&models.BannerModel{},
		&models.UserLoginModel{},
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
	if db.Migrator().HasColumn(&models.LogModel{}, "password") {
		if err := db.Migrator().DropColumn(&models.LogModel{}, "password"); err != nil {
			global.Logger.Errorf("删除日志表 password 列失败: %v", err)
		}
	}
	if db.Migrator().HasTable("image_upload_task_models") {
		if err := db.Migrator().DropTable("image_upload_task_models"); err != nil {
			global.Logger.Errorf("删除旧图片上传任务表失败: %v", err)
		}
	}
	global.Logger.Info("数据库迁移成功")
}
