package article_api

import (
	"errors"
	"fmt"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_article"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleFavoriteRequest struct {
	ArticleID uint `json:"article_id" binding:"required"`
	FavorID   uint `json:"favor_id"`
}

func (ArticleApi) ArticleFavoriteSaveView(c *gin.Context) {
	cr := middleware.GetBindJson[ArticleFavoriteRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	var article models.ArticleModel
	if err := global.DB.Take(&article, "id = ? and status = ?", cr.ArticleID, enum.ArticleStatusPublished).Error; err != nil {
		res.FailWithMsg("查询文章失败", c)
		return
	}

	var isFavorited bool
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		favorite, err := getOrCreateFavoriteID(tx, cr.FavorID, claims.UserID)
		if err != nil {
			return err
		}

		var articleFavorite models.UserArticleFavorModel
		if err = tx.Take(&articleFavorite, "article_id = ? and user_id = ? and favor_id = ?", cr.ArticleID, claims.UserID, favorite.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				articleFavorite = models.UserArticleFavorModel{
					ArticleID: cr.ArticleID,
					UserID:    claims.UserID,
					FavorID:   favorite.ID,
				}
				if err = tx.Create(&articleFavorite).Error; err != nil {
					return err
				}
				if err = tx.Model(&favorite).Update("article_count", gorm.Expr("article_count + 1")).Error; err != nil {
					return err
				}
				isFavorited = true
				return nil
			}
			return err
		}

		if err = tx.Delete(&articleFavorite).Error; err != nil {
			return err
		}
		if err = tx.Model(&favorite).Where("article_count > 0").Update("article_count", gorm.Expr("article_count - 1")).Error; err != nil {
			return err
		}
		isFavorited = false
		return nil
	}); err != nil {
		res.FailWithMsg("收藏操作失败", c)
		return
	}

	if isFavorited {
		redis_article.SetCacheFavorite(cr.ArticleID, 1)
		res.OkWithMsg("收藏成功", c)
	} else {
		redis_article.SetCacheFavorite(cr.ArticleID, -1)
		res.OkWithMsg("取消收藏成功", c)
	}
}

func getOrCreateFavoriteID(db *gorm.DB, favorID, userID uint) (*models.FavoriteModel, error) {
	var favorite models.FavoriteModel
	if favorID == 0 {
		err := db.Take(&favorite, "is_default = ? and user_id = ?", true, userID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				favorite = models.FavoriteModel{
					Title:     "默认收藏夹",
					IsDefault: true,
					UserID:    userID,
				}
				if err := db.Create(&favorite).Error; err != nil {
					return nil, errors.New("创建默认收藏夹失败")
				}
				return &favorite, nil
			}
			return nil, errors.New("查询默认收藏夹失败")
		}
		return &favorite, nil
	}

	if err := db.Take(&favorite, "id = ? and user_id = ?", favorID, userID).Error; err != nil {
		return nil, errors.New("收藏夹不存在")
	}
	return &favorite, nil
}

type FavoriteRequest struct {
	ID       uint   `json:"id"`
	Title    string `json:"title" binding:"required,min=2,max=32"`
	Cover    string `json:"cover"`
	Abstract string `json:"abstract" binding:"required,max=256"`
}

// 创建或者编辑收藏夹（传入ID则视为创建，不传入则视为编辑）
func (ArticleApi) FavoriteCreateUpdateView(c *gin.Context) {
	cr := middleware.GetBindJson[FavoriteRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	// 创建
	if cr.ID == 0 {
		if err := global.DB.Take(&models.FavoriteModel{}, "user_id = ? and title = ?", claims.UserID, cr.Title).Error; err == nil {
			res.FailWithMsg("收藏夹名称重复", c)
			return
		}

		if err := global.DB.Create(&models.FavoriteModel{
			UserID:   claims.UserID,
			Title:    cr.Title,
			Cover:    cr.Cover,
			Abstract: cr.Abstract,
		}).Error; err != nil {
			res.FailWithMsg(fmt.Sprintf("创建收藏夹失败 %v", err), c)
			return
		}
		res.OkWithMsg("创建收藏夹成功", c)
		return
	}

	// 编辑
	var favorite models.FavoriteModel
	if err := global.DB.Take(&favorite, "user_id = ? and id = ?", claims.UserID, cr.ID).Error; err != nil {
		res.FailWithMsg("收藏夹不存在", c)
		return
	}

	if err := global.DB.Model(&favorite).Updates(map[string]any{
		"title":    cr.Title,
		"cover":    cr.Cover,
		"abstract": cr.Abstract,
	}).Error; err != nil {
		res.FailWithMsg(fmt.Sprintf("更新收藏夹失败 %v", err), c)
		return
	}
	res.OkWithMsg("更新收藏夹成功", c)
}

type FavoriteListRequest struct {
	common.PageInfo
	UserID uint `form:"user_id"`
	Type   int8 `form:"type" binding:"required,oneof=1 2 3"` // 1:查自己 2:查别人 3:管理员后台查
}

type FavoriteListResponse struct {
	models.FavoriteModel
	ArticleCount int    `json:"article_count"`
	Nickname     string `json:"nickname,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
}

// 查询收藏夹列表
func (ArticleApi) FavoriteListView(c *gin.Context) {
	cr := middleware.GetBindQuery[FavoriteListRequest](c)

	claim, err := jwts.ParseTokenByGin(c)

	Preloads := []string{"ArticleList"}

	switch cr.Type {
	case 1:
		if err != nil {
			res.FailWithError(err, c)
			return
		}
		cr.UserID = claim.UserID
	case 2: //
	case 3:
		if err != nil || claim.IsAdmin() == false {
			res.FailWithMsg("权限不足", c)
			return
		}
		Preloads = append(Preloads, "UserModel")
	}

	_list, count, _ := common.ListQuery(models.FavoriteModel{
		UserID: cr.UserID,
	}, common.Options{
		PageInfo: cr.PageInfo,
		Likes:    []string{"title"},
		Preloads: Preloads,
	})

	var list = make([]FavoriteListResponse, 0)
	for _, item := range _list {
		list = append(list, FavoriteListResponse{
			FavoriteModel: item,
			ArticleCount:  len(item.ArticleList),
			Nickname:      item.UserModel.Nickname,
			Avatar:        item.UserModel.Avatar,
		})
	}

	res.OkWithList(list, count, c)
}

// 删除收藏夹
func (ArticleApi) FavoriteDeleteView(c *gin.Context) {
	cr := middleware.GetBindJson[models.RemoveRequest](c)

	if len(cr.IDList) == 0 {
		res.FailWithMsg("请填入要删除的 id 列表", c)
		return
	}

	query := global.DB.Where("id IN ?", cr.IDList)

	claim := jwts.GetClaimsByGin(c)
	if claim.IsAdmin() == false {
		query = query.Where("user_id = ?", claim.UserID)
	}

	var list []models.FavoriteModel
	if err := global.DB.Where(query).Find(&list).Error; err != nil {
		global.Logger.Errorf("寻找对应的收藏夹失败 err: %v", err)
		res.FailWithMsg("寻找对应的收藏夹失败", c)
		return
	}

	if len(list) > 0 {
		if err := global.DB.Delete(&list).Error; err != nil {
			global.Logger.Errorf("删除对应的收藏夹失败 err: %v", err)
			res.FailWithMsg("删除收藏夹失败", c)
			return
		}
	} else {
		res.FailWithMsg("未找到需删除的收藏夹", c)
		return
	}

	res.OkWithMsg(fmt.Sprintf("删除收藏夹成功，共删除 %d 条", len(list)), c)
}
