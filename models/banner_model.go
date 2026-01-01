// 轮播图模型

package models

// 轮播图表
type BannerModel struct {
	Model
	Show  bool   `json:"show"`  // 是否显示
	Cover string `json:"cover"` // 封面图片链接
	Href  string `json:"href"`  // 跳转链接
}
