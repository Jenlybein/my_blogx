package category

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/service/log_service"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

// 删除分类
func (CategoryApi) CategoryDeleteView(c *gin.Context) {
	cr := middleware.GetBindJson[models.IDListRequest](c)

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
		global.Logger.Errorf("查找对应分类失败: 错误=%v", err)
		res.FailWithMsg("寻找对应的分类失败", c)
		return
	}

	if len(list) > 0 {
		if err := global.DB.Delete(&list).Error; err != nil {
			global.Logger.Errorf("删除对应分类失败: 错误=%v", err)
			res.FailWithMsg("删除分类失败", c)
			return
		}
	} else {
		res.FailWithMsg("未找到需删除的分类", c)
		return
	}
	res.OkWithMsg(fmt.Sprintf("删除分类成功，共删除 %d 条", len(list)), c)
	log_service.EmitActionAuditFromGin(c, log_service.GinAuditInput{
		ActionName:  "category_delete",
		TargetType:  "category",
		Success:     true,
		Message:     fmt.Sprintf("删除分类成功，共删除 %d 条", len(list)),
		RequestBody: map[string]any{"id_list": cr.IDList},
		ResponseBody: map[string]any{
			"deleted_count": len(list),
		},
		UseRawRequestBody: true,
		UseRawRequestHead: true,
	})
}
