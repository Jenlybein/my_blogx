// API模块入口

package api

import (
	"myblogx/api/banner_api"
	"myblogx/api/image_api"
	"myblogx/api/log_api"
	"myblogx/api/site_api"
)

type Api struct {
	SiteApi   site_api.SiteApi
	LogApi    log_api.LogApi
	ImageApi  image_api.ImageApi
	BannerApi banner_api.BannerApi
}

var App = Api{}
