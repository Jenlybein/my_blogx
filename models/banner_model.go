package models

// 轮播图表
type BannerModel struct {
	Model
	Cover string `json:"cover"` // 封面图片链接
	Herf  string `json:"herf"`  // 跳转链接
}
