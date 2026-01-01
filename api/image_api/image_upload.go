package image_api

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"time"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	"myblogx/common/res"
	"myblogx/global"
	"myblogx/models"
	"myblogx/utils/file"
	"myblogx/utils/hash"
	"myblogx/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (i *ImageApi) ImageUploadView(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 校验图片大小
	if fileHeader.Size > 1024*1024*2 {
		res.FailWithMsg("图片大小不能超过2MB", c)
		return
	}

	// 校验图片格式
	whitelist := global.Config.Upload.Whitelist
	err = file.VerifyImageFormat(whitelist, fileHeader)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 获取文件md5值，判断图片是否重复
	hash, err := hash.FileHeaderMd5(fileHeader)
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	var imageModel models.ImageModel
	err = global.DB.Take(&imageModel, "hash = ?", hash).Error
	if err == nil {
		msg := fmt.Sprintf("图片已存在,文件名:%s,路径:%s", imageModel.FileName, imageModel.Path)
		res.FailWithMsg(msg, c)
		return
	}

	// 文件文件信息
	var username string
	claims, ok := c.Get("claims")
	if ok {
		username = claims.(*jwts.MyClaims).Username
	} else {
		username = "default"
	}
	fileName := fmt.Sprintf(
		"%s_%s.%s",
		username,
		time.Now().Format("20060102150405"),
		file.GetImageSuffix(fileHeader.Filename),
	)
	filePath := fmt.Sprintf("uploads/%s/%s", global.Config.Upload.UploadDir, fileName)

	// 图片入库
	imageModel = models.ImageModel{
		FileName: fileName,
		Path:     filePath,
		Size:     fileHeader.Size,
		Hash:     hash,
	}
	err = global.DB.Create(&imageModel).Error
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	c.SaveUploadedFile(fileHeader, filePath)
	res.Ok(imageModel.WebPath(), "图片上传成功", c)
}
