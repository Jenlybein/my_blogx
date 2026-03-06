package models

import (
	_ "embed"
	"myblogx/global"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_tag"

	"gorm.io/gorm"
)

// ArticleModel 文章表
type ArticleModel struct {
	Model
	Title          string             `gorm:"size:256" json:"title"`
	Abstract       string             `gorm:"size:256" json:"abstract"`
	Content        string             `gorm:"type:longtext" json:"content"`
	HtmlContent    string             `gorm:"type:longtext" json:"html_content"`
	CategoryID     *uint              `gorm:"index" json:"category_id"`
	Cover          string             `gorm:"size:256" json:"cover"`
	AuthorID       uint               `gorm:"index" json:"author_id"`
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
	tx.Find(&commentList, "article_id = ?", a.ID).Delete(&commentList)

	var diggList []ArticleDiggModel
	tx.Find(&diggList, "article_id = ?", a.ID).Delete(&diggList)

	var favoriteList []FavoriteModel
	tx.Find(&favoriteList, "article_id = ?", a.ID).Delete(&favoriteList)

	var topList []UserTopArticleModel
	tx.Find(&topList, "article_id = ?", a.ID).Delete(&topList)

	var viewList []UserArticleViewHistoryModel
	tx.Find(&viewList, "article_id = ?", a.ID).Delete(&viewList)

	var articleTagList []ArticleTagModel
	tx.Find(&articleTagList, "article_id = ?", a.ID).Delete(&articleTagList)
	if global.Redis != nil {
		for _, relation := range articleTagList {
			if cacheErr := redis_tag.SetCacheArticleCount(relation.TagID, -1); cacheErr != nil {
				global.Logger.Errorf("标签文章数缓存减少失败 tag_id=%d err=%v", relation.TagID, cacheErr)
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
