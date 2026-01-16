package file

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"myblogx/global"
	"path/filepath"
	"strings"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	"gorm.io/gorm/utils"

	"github.com/gabriel-vasile/mimetype"
)

// 获取图片后缀（不包含前缀点）
func GetImageSuffix(str string) string {
	suffix := strings.ToLower(filepath.Ext(str))
	if strings.HasPrefix(suffix, ".") {
		return suffix[1:]
	}
	return ""
}

// 校验图片格式
func VerifyImageFormat(whitelist []string, fileHeader *multipart.FileHeader) error {
	// 校验后缀是否在白名单中
	suffix := GetImageSuffix(fileHeader.Filename)
	if suffix == "" {
		return errors.New("图片名中的格式后缀错误")
	}
	if !utils.Contains(whitelist, suffix) {
		return fmt.Errorf("图片后缀 %s 不在服务器允许上传的图片格式白名单中", suffix)
	}

	// 创建读取器
	file, err := fileHeader.Open()
	if err != nil {
		serr := fmt.Errorf("图片格式验证时，创建文件读取器失败：%w", err)
		global.Logger.Error(serr)
		return serr
	}
	defer file.Close()

	// 实现可回退读指针，用于多次读取，避免直接读取导致文件指针移动
	rs, ok := file.(io.ReadSeeker)
	if !ok {
		return errors.New("上传的文件无法进行完整校验")
	}

	// 创建 MIME 类型检测器
	mt, err := mimetype.DetectReader(rs)
	if err != nil {
		return err
	}

	// MIME 类型检测（判断是否为真实图片，避免上传非图片文件）
	if suffix == "jpg" {
		suffix = "jpeg"
	}
	mime := fmt.Sprintf("image/%s", suffix)
	if !mt.Is(mime) {
		return errors.New("非服务器允许上传的图片格式")
	}

	// DetectReader 会消耗 reader，需要复位文件指针
	if _, err = rs.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// 图片结构校验（判断能否被当成图片解析）
	cfg, _, err := image.DecodeConfig(rs)
	if err != nil {
		return fmt.Errorf("图片结构校验失败：%w", err)
	}

	// 尺寸限制，防图片炸弹
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 10000 || cfg.Height > 10000 {
		return errors.New("图片尺寸不合规或图片尺寸过大")
	}

	return nil
}
