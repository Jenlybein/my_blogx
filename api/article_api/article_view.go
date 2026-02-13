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
	ctx := c.Request.Context()
	now := time.Now()

	// 验证文章是否存在并已发布
	var articleID uint
	err := global.DB.Model(&models.ArticleModel{}).
		Where("id = ? and status = ?", cr.ArticleID, enum.ArticleStatusPublished).
		Select("id").Scan(&articleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.FailWithMsg("文章不存在或未发布", c)
			return
		}
		// 记录详细错误日志（建议使用日志库，如 zap）
		global.Logger.Errorf("数据库验证文章失败 %v, article_id: %d", err, cr.ArticleID)
		res.FailWithMsg("服务器内部错误", c)
		return
	}

	claims, err := jwts.ParseTokenByGin(c)

	var keySuffix string

	if claims == nil {
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
		keySuffix = fmt.Sprintf("g:%s", hex.EncodeToString(hash[:]))
	} else {
		// 已登录用户，靠用户 id 进行确认
		keySuffix = fmt.Sprintf("u:%d", claims.UserID)

		// 同时更新数据库浏览历史(TODO：改消息队列)
		condition := map[string]interface{}{
			"article_id": cr.ArticleID,
			"user_id":    claims.UserID,
		}
		var articleHistory models.UserArticleViewHistoryModel
		if dbErr := global.DB.Where(condition).FirstOrCreate(&articleHistory, condition).Error; dbErr != nil {
			global.Logger.Error("查询/创建文章访问历史失败", "err", dbErr, "article_id", cr.ArticleID, "user_id", claims.UserID)
			res.FailWithMsg("服务器内部错误", c)
			return
		} else {
			if dbErr := global.DB.Model(&articleHistory).Update("updated_at", now).Error; dbErr != nil {
				global.Logger.Error("更新文章访问历史时间失败", "err", dbErr, "article_id", cr.ArticleID, "user_id", claims.UserID)
				res.FailWithMsg("服务器内部错误", c)
				return
			}
		}
	}

	// 将用户id和文章id作为key存入缓存，value为访问时间，过期时间24小时
	cacheKey := fmt.Sprintf("article_visit:%s:%d", keySuffix, cr.ArticleID)
	hashKey := fmt.Sprintf("%s:%s", string(redis_article.ArticleCacheGuestView), now.Format("2006-01-02"))
	nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	// 使用 Redis Pipeline 保证多操作原子性（批量执行，要么都成功，要么都失败）
	pipe := global.Redis.Pipeline()
	hsetnxCmd := pipe.HSetNX(ctx, hashKey, cacheKey, now)
	pipe.ExpireAt(ctx, hashKey, nextDay)

	_, pipeErr := pipe.Exec(ctx)
	if pipeErr != nil {
		global.Logger.Errorf("Redis 操作失败 %v, cacheKey: %s", pipeErr, cacheKey)
		res.FailWithMsg("访问记录统计失败", c)
		return
	}

	// 首次访问则返回true，自增访问量
	if hsetnxCmd.Val() {
		redis_article.SetCacheView(cr.ArticleID, 1)
		res.OkWithMsg("文章访问量增加成功", c)
	} else {
		res.OkWithMsg("用户已访问过该文章", c)
	}

}
