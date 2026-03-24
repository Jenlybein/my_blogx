package category

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (CategoryApi) CategoryOptionsView(c *gin.Context) {

	claims := jwts.GetClaimsByGin(c)

	var list []models.OptionsResponse[ctype.ID]
	if err := global.DB.Model(&models.CategoryModel{}).Where("user_id = ?", claims.UserID).Select("id as value", "title as label").Scan(&list).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	res.OkWithData(list, c)
}
