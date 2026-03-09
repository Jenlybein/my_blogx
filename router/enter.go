// 路由模块入口

package router

import (
	"myblogx/global"
	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func Run() {
	gin.SetMode(global.Config.System.GinMode)
	r := gin.Default()

	r.Static("/uploads", "./uploads")

	nr := r.Group("/api")
	nr.Use(middleware.LogMiddleware)

	SiteRouter(nr)
	LogRouter(nr)
	ImageRouter(nr)
	BannerRouter(nr)
	CaptchaRouter(nr)
	UserRouter(nr)
	ArticleRouter(nr)
	CommentRouter(nr)
	SitemsgRouter(nr)
	GlobalNotifRouter(nr)

	addr := global.Config.System.Addr()
	r.Run(addr)
}
