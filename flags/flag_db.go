package flags

import (
	"myblogx/global"
	"myblogx/models"

	"github.com/sirupsen/logrus"
)

func FlagDB() {
	// 确保数据库使用 utf8mb4 字符集
	global.DB.Exec("SET character_set_client = utf8mb4")
	global.DB.Exec("SET character_set_connection = utf8mb4")
	global.DB.Exec("SET character_set_database = utf8mb4")
	global.DB.Exec("SET character_set_results = utf8mb4")
	global.DB.Exec("SET character_set_server = utf8mb4")
	global.DB.Exec("SET collation_connection = utf8mb4_unicode_ci")
	global.DB.Exec("SET collation_database = utf8mb4_unicode_ci")
	global.DB.Exec("SET collation_server = utf8mb4_unicode_ci")

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
