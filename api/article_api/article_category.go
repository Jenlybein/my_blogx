package article_api

import (
	"fmt"
	"myblogx/common"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

type CategoryRequest struct {
	ID    uint   `json:"id"`
	Title string `json:"title" binding:"required,min=2,max=20"`
}

// 创建或者编辑分类（传入ID则视为创建，不传入则视为编辑）
func (ArticleApi) CategoryCreateUpdateView(c *gin.Context) {
	cr := middleware.GetBindJson[CategoryRequest](c)
	claims := jwts.MustGetClaimsByGin(c)

	// 创建
	if cr.ID == 0 {
		if err := global.DB.Take(&models.CategoryModel{}, "user_id = ? and title = ?", claims.UserID, cr.Title).Error; err == nil {
			res.FailWithMsg("分类名称重复", c)
			return
		}

		if err := global.DB.Create(&models.CategoryModel{
			Title:  cr.Title,
			UserID: claims.UserID,
		}).Error; err != nil {
			res.FailWithMsg(fmt.Sprintf("创建分类失败 %v", err), c)
			return
		}
		res.OkWithMsg("创建成功", c)
		return
	}

	// 编辑
	var category models.CategoryModel
	if err := global.DB.Take(&category, "user_id = ? and id = ?", claims.UserID, cr.ID).Error; err != nil {
		res.FailWithMsg("分类不存在", c)
		return
	}

	if err := global.DB.Model(&category).Update("title", cr.Title).Error; err != nil {
		res.FailWithMsg(fmt.Sprintf("更新分类失败 %v", err), c)
		return
	}
	res.OkWithMsg("更新分类成功", c)
}

type CategoryListRequest struct {
	common.PageInfo
	UserID uint `form:"user_id"`
	Type   int8 `form:"type" binding:"required,oneof=1 2 3"` // 1:查自己 2:查别人 3:管理员后台查
}

type CategoryListResponse struct {
	models.CategoryModel
	ArticleCount int    `json:"article_count"`
	Nickname     string `json:"nickname,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
}

// 查询分类列表
func (ArticleApi) CategoryListView(c *gin.Context) {
	cr := middleware.GetBindQuery[CategoryListRequest](c)

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

	_list, count, _ := common.ListQuery(models.CategoryModel{
		UserID: cr.UserID,
	}, common.Options{
		PageInfo: cr.PageInfo,
		Likes:    []string{"title"},
		Preloads: Preloads,
	})

	var list = make([]CategoryListResponse, 0)
	for _, item := range _list {
		list = append(list, CategoryListResponse{
			CategoryModel: item,
			ArticleCount:  len(item.ArticleList),
			Nickname:      item.UserModel.Nickname,
			Avatar:        item.UserModel.Avatar,
		})
	}

	res.OkWithList(list, count, c)
}

// 删除分类
func (ArticleApi) CategoryDeleteView(c *gin.Context) {
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

	var list []models.CategoryModel
	if err := global.DB.Where(query).Find(&list).Error; err != nil {
		global.Logger.Errorf("寻找对应的分类失败 err: %v", err)
		res.FailWithMsg("寻找对应的分类失败", c)
		return
	}

	if len(list) > 0 {
		if err := global.DB.Delete(&list).Error; err != nil {
			global.Logger.Errorf("删除对应的分类失败 err: %v", err)
			res.FailWithMsg("删除分类失败", c)
			return
		}
	} else {
		res.FailWithMsg("未找到需删除的分类", c)
		return
	}

	res.OkWithMsg(fmt.Sprintf("删除分类成功，共删除 %d 条", len(list)), c)
}
