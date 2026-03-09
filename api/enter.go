// API模块入口

package api

import (
	"myblogx/api/article_api"
	"myblogx/api/banner_api"
	"myblogx/api/captcha_api"
	"myblogx/api/comment_api"
	"myblogx/api/image_api"
	"myblogx/api/log_api"
	"myblogx/api/site_api"
	"myblogx/api/sitemsg_api"
	"myblogx/api/user_api"
)

type Api struct {
	SiteApi         site_api.SiteApi
	LogApi          log_api.LogApi
	ImageApi        image_api.ImageApi
	BannerApi       banner_api.BannerApi
	ImageCaptchaApi captcha_api.ImageCaptchaApi
	UserApi         user_api.UserApi
	ArticleApi      article_api.ArticleApi
	CommentApi      comment_api.CommentApi
	SitemsgApi      sitemsg_api.SitemsgApi
}

var App = Api{}
