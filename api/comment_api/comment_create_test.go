package comment_api

import (
	"encoding/json"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_article"
	"myblogx/service/redis_service/redis_comment"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCommentCtx() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func readBizBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v body=%s", err, w.Body.String())
	}
	return body
}

func readBizCode(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	body := readBizBody(t, w)
	return int(body["code"].(float64))
}

func setupCommentEnv(t *testing.T) *models.UserModel {
	t.Helper()
	_ = testutil.SetupMiniRedis(t)
	db := testutil.SetupSQLite(t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.UserFollowModel{},
		&models.ArticleModel{},
		&models.CommentModel{},
		&models.CommentDiggModel{},
		&models.ArticleMessageModel{},
	)
	global.Config.Site.Comment.SkipExamining = true

	user := &models.UserModel{
		Username: "comment_u",
		Password: "x",
		Role:     enum.RoleUser,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	return user
}

func TestCommentCreateView(t *testing.T) {
	user := setupCommentEnv(t)
	api := CommentApi{}
	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Username: user.Username, Role: enum.RoleUser}}

	t.Run("文章不存在", func(t *testing.T) {
		c, w := newCommentCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CommentCreateRequest{ArticleID: 999, Content: "x"})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code == 0 {
			t.Fatalf("文章不存在应失败, body=%s", w.Body.String())
		}
	})

	closedArticle := models.ArticleModel{
		Title:          "closed",
		Content:        "c",
		AuthorID:       user.ID,
		CommentsToggle: true,
	}
	if err := global.DB.Create(&closedArticle).Error; err != nil {
		t.Fatalf("创建关闭评论文章失败: %v", err)
	}
	if err := global.DB.Model(&closedArticle).Update("comments_toggle", false).Error; err != nil {
		t.Fatalf("关闭评论开关失败: %v", err)
	}

	t.Run("文章关闭评论", func(t *testing.T) {
		c, w := newCommentCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CommentCreateRequest{ArticleID: closedArticle.ID, Content: "x"})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code == 0 {
			t.Fatalf("关闭评论应失败, body=%s", w.Body.String())
		}
	})

	openArticle := models.ArticleModel{
		Title:          "open",
		Content:        "c",
		AuthorID:       user.ID,
		CommentsToggle: true,
	}
	if err := global.DB.Create(&openArticle).Error; err != nil {
		t.Fatalf("创建可评论文章失败: %v", err)
	}

	var first models.CommentModel
	t.Run("一级评论成功并写入缓存与消息", func(t *testing.T) {
		c, w := newCommentCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CommentCreateRequest{ArticleID: openArticle.ID, Content: "first"})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code != 0 {
			t.Fatalf("一级评论应成功, body=%s", w.Body.String())
		}

		if err := global.DB.Last(&first).Error; err != nil {
			t.Fatalf("查询评论失败: %v", err)
		}
		if first.ReplyId != 0 || first.RootID != 0 {
			t.Fatalf("一级评论 reply/root 错误: %+v", first)
		}
		if first.Status != enum.CommentStatusPublished {
			t.Fatalf("一级评论状态错误: %d", first.Status)
		}
		if redis_article.GetCacheComment(openArticle.ID) != 1 {
			t.Fatalf("评论缓存计数错误: %d", redis_article.GetCacheComment(openArticle.ID))
		}
		if redis_comment.GetCacheReply(first.ID) != 0 {
			t.Fatalf("一级评论回复缓存初始值应为 0: %d", redis_comment.GetCacheReply(first.ID))
		}
	})

	var second models.CommentModel
	t.Run("回复一级评论成功并累加缓存与消息", func(t *testing.T) {
		c, w := newCommentCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CommentCreateRequest{
			ArticleID: openArticle.ID,
			Content:   "reply-level2",
			ReplyId:   &first.ID,
		})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code != 0 {
			t.Fatalf("回复评论应成功, body=%s", w.Body.String())
		}

		if err := global.DB.Last(&second).Error; err != nil {
			t.Fatalf("查询回复评论失败: %v", err)
		}
		if second.ReplyId != first.ID {
			t.Fatalf("reply_id 错误: %+v", second)
		}
		if second.RootID != first.ID {
			t.Fatalf("root_id 错误: %+v", second)
		}
		if second.Status != enum.CommentStatusPublished {
			t.Fatalf("回复评论状态错误: %d", second.Status)
		}
		if redis_article.GetCacheComment(openArticle.ID) != 2 {
			t.Fatalf("评论缓存计数错误: %d", redis_article.GetCacheComment(openArticle.ID))
		}
		if redis_comment.GetCacheReply(first.ID) != 1 {
			t.Fatalf("一级评论 ReplyCount 缓存错误: %d", redis_comment.GetCacheReply(first.ID))
		}
	})

	t.Run("回复二级评论仍归属同一一级评论", func(t *testing.T) {
		c, w := newCommentCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CommentCreateRequest{
			ArticleID: openArticle.ID,
			Content:   "reply-level2-again",
			ReplyId:   &second.ID,
		})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code != 0 {
			t.Fatalf("回复二级评论应成功, body=%s", w.Body.String())
		}

		var reply models.CommentModel
		if err := global.DB.Last(&reply).Error; err != nil {
			t.Fatalf("查询回复评论失败: %v", err)
		}
		if reply.ReplyId != second.ID {
			t.Fatalf("reply_id 错误: %+v", reply)
		}
		if reply.RootID != first.ID {
			t.Fatalf("root_id 应保持一级评论 ID: %+v", reply)
		}
		if redis_comment.GetCacheReply(first.ID) != 2 {
			t.Fatalf("一级评论 ReplyCount 缓存错误: %d", redis_comment.GetCacheReply(first.ID))
		}
	})

	t.Run("回复评论不存在", func(t *testing.T) {
		missing := ctype.ID(123456)
		c, w := newCommentCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CommentCreateRequest{
			ArticleID: openArticle.ID,
			Content:   "bad",
			ReplyId:   &missing,
		})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code == 0 {
			t.Fatalf("回复评论不存在应失败 body=%s", w.Body.String())
		}
	})

	t.Run("关闭免审核时普通用户评论进入审核中且不计数", func(t *testing.T) {
		global.Config.Site.Comment.SkipExamining = false
		t.Cleanup(func() {
			global.Config.Site.Comment.SkipExamining = true
		})

		beforeCommentCount := redis_article.GetCacheComment(openArticle.ID)
		beforeReplyCount := redis_comment.GetCacheReply(first.ID)

		c, w := newCommentCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CommentCreateRequest{
			ArticleID: openArticle.ID,
			Content:   "need-examining",
			ReplyId:   &first.ID,
		})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code != 0 {
			t.Fatalf("评论提交应成功, body=%s", w.Body.String())
		}

		var last models.CommentModel
		if err := global.DB.Last(&last).Error; err != nil {
			t.Fatalf("查询评论失败: %v", err)
		}
		if last.Status != enum.CommentStatusExamining {
			t.Fatalf("评论状态应为审核中: %d", last.Status)
		}
		if redis_article.GetCacheComment(openArticle.ID) != beforeCommentCount {
			t.Fatalf("审核中评论不应增加文章评论缓存")
		}
		if redis_comment.GetCacheReply(first.ID) != beforeReplyCount {
			t.Fatalf("审核中评论不应增加回复缓存")
		}
	})

	t.Run("管理员评论直接发布并计数与消息", func(t *testing.T) {
		global.Config.Site.Comment.SkipExamining = false
		adminClaims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Username: user.Username, Role: enum.RoleAdmin}}

		beforeCommentCount := redis_article.GetCacheComment(openArticle.ID)
		beforeReplyCount := redis_comment.GetCacheReply(first.ID)

		c, w := newCommentCtx()
		c.Set("claims", adminClaims)
		c.Set("requestJson", CommentCreateRequest{
			ArticleID: openArticle.ID,
			Content:   "admin-pass",
			ReplyId:   &first.ID,
		})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code != 0 {
			t.Fatalf("管理员评论应成功 body=%s", w.Body.String())
		}

		var last models.CommentModel
		if err := global.DB.Last(&last).Error; err != nil {
			t.Fatalf("查询评论失败: %v", err)
		}
		if last.Status != enum.CommentStatusPublished {
			t.Fatalf("管理员评论应直接发布: %d", last.Status)
		}
		if redis_article.GetCacheComment(openArticle.ID) != beforeCommentCount+1 {
			t.Fatalf("管理员评论应增加文章评论缓存")
		}
		if redis_comment.GetCacheReply(first.ID) != beforeReplyCount+1 {
			t.Fatalf("管理员评论应增加回复缓存")
		}
	})
}
