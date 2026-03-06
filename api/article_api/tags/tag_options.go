package tags

import (
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"

	"github.com/gin-gonic/gin"
)

func (TagsApi) ArticleTagOptionsView(c *gin.Context) {
	var list []models.OptionsResponse[uint]
	if err := global.DB.Model(&models.TagModel{}).
		Where("is_enabled = ?", true).
		Order("sort desc, id asc").
		Select("id as value", "title as label").
		Scan(&list).Error; err != nil {
		res.FailWithError(err, c)
		return
	}

	res.OkWithData(list, c)
}
