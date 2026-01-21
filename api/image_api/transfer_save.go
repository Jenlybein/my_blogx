package image_api

import (
	"fmt"
	"io"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/utils/hash"
	"net/http"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

type TransferSaveRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

// 图片转存到本站，格式用png（保证图片显示效果）
func (i *ImageApi) TransferSaveView(c *gin.Context) {
	cr := middleware.GetBindJson[TransferSaveRequest](c)

	resp, err := http.Get(cr.ImageURL)
	if err != nil {
		res.FailWithMsg("获取图片失败："+err.Error(), c)
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		res.FailWithMsg("读取图片失败："+err.Error(), c)
		return
	}

	hash := hash.Md5(body)

	fileName := fmt.Sprintf("%s%s", hash, mimetype.Detect(body).Extension())
	filePath := fmt.Sprintf("uploads/%s/%s", global.Config.Upload.UploadDir, fileName)

	err = os.WriteFile(filePath, body, 0644)
	if err != nil {
		res.FailWithMsg("写入图片失败："+err.Error(), c)
		return
	}
	res.Ok("/"+filePath, "图片上传成功", c)
}
