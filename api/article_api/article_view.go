// 该接口与文章内容返回接口分开，加快文章内容返回速度

package article_api

import (
	"errors"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/middleware"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_article"
	"myblogx/utils/jwts"
	"myblogx/utils/user_info"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleViewCountRequest struct {
	ArticleID uint `json:"article_id" binding:"required"`
}

func (ArticleApi) ArticleVisitView(c *gin.Context) {
	cr := middleware.GetBindJson[ArticleViewCountRequest](c)

	// TODO：引入缓存，并发处理
	// 将用户id和文章id作为key存入缓存，value为访问时间，过期时间24小时
	// 如果缓存中已存在该key，说明用户已访问过，不重复统计，但更新访问历史
	// 如果缓存中不存在该key，说明用户未访问过，统计浏览量，并将key-value存入缓存

	var article models.ArticleModel
	if err := global.DB.Take(&article, "id = ? and status = ?", cr.ArticleID, enum.ArticleStatusPublished).Error; err != nil {
		res.FailWithMsg("查询文章失败", c)
		return
	}

	claims, err := jwts.ParseTokenByGin(c)

	if err != nil {
		// 未登录用户，靠 ip 和 设备id 进行确认
		ip := user_info.GetClientIP(c)
		if ip == "" {
			res.OkWithMsg("无法获取IP，跳过统计", c)
			return
		}
		ua := c.GetHeader("User-Agent")
		if ua == "" {
			ua = "unknown"
		}
		if len(ua) > 255 {
			ua = ua[:255] // 截断过长的User-Agent，避免超过数据库字段限制
		}
		// TODO：获取更真实可靠的ip和设备id防爬虫？
		if err = global.DB.Transaction(func(tx *gorm.DB) error {
			// 查询访客今天是否已访问过该文章
			var guestArticleViewRecord models.GuestArticleViewRecordModel
			if err = tx.Take(&guestArticleViewRecord,
				"article_id = ? and guest_ip = ? and device_id = ? and created_at >= ?",
				cr.ArticleID, ip, ua,
				time.Now().Truncate(24*time.Hour),
			).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 访客今天未访问过，创建记录
					guestArticleViewRecord = models.GuestArticleViewRecordModel{
						ArticleID: cr.ArticleID,
						GuestIP:   ip,
						DeviceID:  ua,
						CreatedAt: time.Now(),
					}
					if err = tx.Create(&guestArticleViewRecord).Error; err != nil {
						return errors.New("创建访客访问记录失败")
					}
					redis_article.SetCacheView(cr.ArticleID, 1)
					return nil
				}
				return errors.New("查询访客访问记录失败")
			}
			return nil
		}); err != nil {
			res.FailWithMsg(err.Error(), c)
			return
		}
		redis_article.SetCacheView(cr.ArticleID, 1)
		res.OkWithMsg("文章访问量增加：访客", c)
		return
	}

	// 已登录用户，检查是否已访问过
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		var articleHistory models.UserArticleViewHistoryModel
		if err = tx.Take(&articleHistory,
			"article_id = ? and user_id = ? and created_at >= ?",
			cr.ArticleID, claims.UserID,
			time.Now().Truncate(24*time.Hour),
		).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				articleHistory = models.UserArticleViewHistoryModel{
					ArticleID: cr.ArticleID,
					UserID:    claims.UserID,
				}
				if err = tx.Create(&articleHistory).Error; err != nil {
					return errors.New("创建访问记录失败")
				}
				if err = tx.Model(&article).Update("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
					return errors.New("更新浏览量失败")
				}
				return nil
			}
			return errors.New("查询访问记录失败")
		}
		// 用户已访问过，不重复统计，但更新访问历史
		if err = tx.Model(&articleHistory).Update("updated_at", time.Now()).Error; err != nil {
			return errors.New("更新访问记录时间失败")
		}
		return nil
	})

	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	redis_article.SetCacheView(cr.ArticleID, 1)

	res.OkWithMsg("文章访问量增加：用户", c)
}
