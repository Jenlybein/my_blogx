// 收藏模型

package models

import (
	"myblogx/global"
	"myblogx/models/ctype"
	"myblogx/service/redis_service/redis_article"

	"gorm.io/gorm"
)

// 收藏表
type FavoriteModel struct {
	Model
	UserID      ctype.ID                `gorm:"uniqueIndex:uk_favorite_user_title,priority:1;index" json:"user_id"`
	Title       string                  `gorm:"size:32;uniqueIndex:uk_favorite_user_title,priority:2" json:"title"`
	Cover       string                  `gorm:"size:256" json:"cover"`
	Abstract    string                  `gorm:"size:256" json:"abstract"`
	IsDefault   bool                    `gorm:"default:false" json:"is_default"`
	UserModel   UserModel               `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleList []UserArticleFavorModel `gorm:"foreignKey:FavorID" json:"-"`
	// ArticleCount int                     `gorm:"default:0" json:"article_count"`
}

func (f *FavoriteModel) BeforeDelete(tx *gorm.DB) (err error) {
	var favorList []UserArticleFavorModel
	if err = tx.Find(&favorList, "favor_id = ?", f.ID).Error; err != nil {
		return err
	}

	if err = tx.Delete(&favorList).Error; err != nil {
		return err
	}

	for _, favor := range favorList {
		if err = redis_article.SetCacheFavorite(favor.ArticleID, -1); err != nil {
			global.Logger.Errorf("文章收藏数据减一失败: 错误=%v", err)
		}
	}

	return nil
}
