// API模块入口

package api

import (
	"myblogx/api/log_api"
	"myblogx/api/site_api"
)

type Api struct {
	SiteApi site_api.SiteApi
	LogApi  log_api.LogApi
}

var App = Api{}
