package flags

import (
	"myblogx/global"
	"myblogx/models"

	"github.com/sirupsen/logrus"
)

func FlagDB() {
	// 自动建表
	err := global.DB.AutoMigrate(
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ArticleModel{},
		&models.ArticleDiggModel{},
		&models.CategoryModel{},
		&models.FavorModel{},
		&models.UserArticleFavorModel{},
		&models.UserArticleViewHistoryModel{},
		&models.UserTopArticleModel{},
		&models.ImageModel{},
		&models.CommentModel{},
		&models.LogModel{},
		&models.BannerModel{},
		&models.UserLoginModel{},
		&models.GlobalNotificationModel{},
	)
	if err != nil {
		logrus.Error("数据库迁移失败: ", err)
		return
	}
	logrus.Info("数据库迁移成功")
}
