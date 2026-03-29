package models

import (
	_ "embed"
	"myblogx/global"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_tag"

	"gorm.io/gorm"
)

// ESTag 是文章写入 ES 时使用的标签结构。
// 这里保留标签 id 和 title，方便 ES 侧按单字段过滤，也方便列表展示。
type ESTag struct {
	ID    ctype.ID `json:"id"`
	Title string   `json:"title"`
}

// ArticleModel 文章表
type ArticleModel struct {
	Model
	Title          string             `gorm:"size:256" json:"title"`
	Abstract       string             `gorm:"size:256" json:"abstract"`
	Content        string             `gorm:"type:longtext" json:"content"`
	ContentHead    string             `gorm:"-" json:"content_head,omitempty"` // ES 冗余字段，保存去除 Markdown 格式后的正文前 150 字
	CategoryID     *ctype.ID          `gorm:"index" json:"category_id"`
	Cover          string             `gorm:"size:256" json:"cover"`
	AuthorID       ctype.ID           `gorm:"index" json:"author_id"`
	ViewCount      int                `gorm:"default:0" json:"view_count"`
	DiggCount      int                `gorm:"default:0" json:"digg_count"`
	CommentCount   int                `gorm:"default:0" json:"comment_count"`
	FavorCount     int                `gorm:"default:0" json:"favor_count"`
	CommentsToggle bool               `gorm:"default:true" json:"comments_toggle"`
	Status         enum.ArticleStatus `gorm:"default:0" json:"status"`
	UserModel      UserModel          `gorm:"foreignKey:AuthorID;references:ID" json:"-"`
	CategoryModel  *CategoryModel     `gorm:"foreignKey:CategoryID;references:ID" json:"-"`
	Tags           []TagModel         `gorm:"many2many:article_tag_models;joinForeignKey:ArticleID;joinReferences:TagID" json:"tags"`
}

//go:embed es_settings/article_mapping.json
var ArticleMapping string

func (ArticleModel) Mapping() string {
	return ArticleMapping
}

func (ArticleModel) Index() string {
	return global.Config.ES.Index
}

//go:embed es_settings/article_pipeline.json
var ArticlePipeline string

func (ArticleModel) Pipeline() string {
	return ArticlePipeline
}

func (ArticleModel) PipelineName() string {
	return "article_pipeline"
}

func (a *ArticleModel) BeforeDelete(tx *gorm.DB) (err error) {
	var commentList []CommentModel
	if err = tx.Where("article_id = ?", a.ID).Find(&commentList).Error; err != nil {
		return err
	}
	if err = tx.Where("article_id = ?", a.ID).Delete(&CommentModel{}).Error; err != nil {
		return err
	}

	var diggList []ArticleDiggModel
	if err = tx.Where("article_id = ?", a.ID).Find(&diggList).Error; err != nil {
		return err
	}
	if err = tx.Where("article_id = ?", a.ID).Delete(&ArticleDiggModel{}).Error; err != nil {
		return err
	}

	var favoriteList []UserArticleFavorModel
	if err = tx.Where("article_id = ?", a.ID).Find(&favoriteList).Error; err != nil {
		return err
	}
	if err = tx.Where("article_id = ?", a.ID).Delete(&UserArticleFavorModel{}).Error; err != nil {
		return err
	}

	var topList []UserTopArticleModel
	if err = tx.Where("article_id = ?", a.ID).Find(&topList).Error; err != nil {
		return err
	}
	if err = tx.Where("article_id = ?", a.ID).Delete(&UserTopArticleModel{}).Error; err != nil {
		return err
	}

	var viewList []UserArticleViewHistoryModel
	if err = tx.Where("article_id = ?", a.ID).Find(&viewList).Error; err != nil {
		return err
	}
	if err = tx.Where("article_id = ?", a.ID).Delete(&UserArticleViewHistoryModel{}).Error; err != nil {
		return err
	}

	var articleTagList []ArticleTagModel
	if err = tx.Where("article_id = ?", a.ID).Find(&articleTagList).Error; err != nil {
		return err
	}
	if err = tx.Where("article_id = ?", a.ID).Delete(&ArticleTagModel{}).Error; err != nil {
		return err
	}
	if global.Redis != nil {
		for _, relation := range articleTagList {
			if cacheErr := redis_tag.SetCacheArticleCount(relation.TagID, -1); cacheErr != nil {
				global.Logger.Errorf("标签文章数缓存减少失败: 标签ID=%d 错误=%v", relation.TagID, cacheErr)
			}
		}
	}

	global.Logger.Infof(
		"删除文章 %d 时，删除了 %d 条评论、%d 条点赞、%d 条收藏、%d 条置顶、%d 条浏览记录、%d 条标签关系",
		a.ID,
		len(commentList),
		len(diggList),
		len(favoriteList),
		len(topList),
		len(viewList),
		len(articleTagList),
	)

	return nil
}
