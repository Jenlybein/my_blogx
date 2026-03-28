package core

import (
	"myblogx/service/site_service"
)

func InitRuntimeSite() error {
	return site_service.InitRuntimeConfig()
}
