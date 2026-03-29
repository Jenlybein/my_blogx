// 该接口与文章内容返回接口分开，加快文章内容返回速度

package article_api

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_article"
	"myblogx/service/user_service"
	"myblogx/utils/user_info"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (ArticleApi) ArticleVisitView(c *gin.Context) {
	cr := middleware.GetBindJson[ArticleViewCountRequest](c)

	// 获取用户登录信息
	authResult := user_service.MustAuthenticateAccessTokenByGin(c)

	if authResult == nil {
		// TODO：获取更真实可靠的ip和设备id防爬虫？
		// 未登录用户，靠 ip 和 设备id 进行确认
		ip := user_info.GetClientIP(c)
		ua := c.GetHeader("User-Agent")
		if ip == "" || ua == "" {
			res.OkWithMsg("无法获取有效访问标识，跳过统计", c)
			return
		}

		// 先生成 ip:ua 字符串，再转为字节切片计算 MD5
		hash := md5.Sum([]byte(fmt.Sprintf("%s:%s", ip, ua)))
		key := fmt.Sprintf("g:%s", hex.EncodeToString(hash[:]))

		if redis_article.GetGuestArticleHistoryCache(int(cr.ArticleID), key) {
			fmt.Printf("访客已经阅读过该文章, %d", cr.ArticleID)
			res.OkWithMsg("访客已访问过该文章", c)
			return
		}

		redis_article.SetGuestArticleHistoryCache(int(cr.ArticleID), key)
	} else {
		claims := authResult.Claims
		// 已登录用户，靠用户 id 进行确认
		if redis_article.GetUserArticleHistoryCache(int(cr.ArticleID), int(claims.UserID)) {
			// TODO：加消息队列通知数据库更新访问历史
			res.OkWithMsg("用户已访问过该文章", c)
			return
		}

		// 验证文章是否存在并已发布
		var articleID ctype.ID
		err := global.DB.Model(&models.ArticleModel{}).
			Where("id = ? and status = ?", cr.ArticleID, enum.ArticleStatusPublished).
			Select("id").Take(&articleID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				res.FailWithMsg("文章不存在或未发布", c)
				return
			}
			// 记录详细错误日志（建议使用日志库，如 zap）
			global.Logger.Errorf("数据库验证文章失败: 错误=%v 文章ID=%d", err, cr.ArticleID)
			res.FailWithMsg("服务器内部错误", c)
			return
		}

		// 同时更新数据库浏览历史(TODO：可选改消息队列异步)
		articleHistory := models.UserArticleViewHistoryModel{
			ArticleID: cr.ArticleID,
			UserID:    claims.UserID,
		}

		if err = global.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "article_id"},
				{Name: "user_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"updated_at": time.Now(),
				"deleted_at": nil,
			}),
		}).Create(&articleHistory).Error; err != nil {
			global.Logger.Errorf("数据库更新浏览历史失败: 错误=%v 文章ID=%d", err, cr.ArticleID)
			res.FailWithMsg("服务器内部错误", c)
			return
		}

		redis_article.SetUserArticleHistoryCache(int(cr.ArticleID), int(claims.UserID))
	}

	redis_article.SetCacheView(cr.ArticleID, 1)
	res.OkWithMsg("文章访问量增加成功", c)
}
