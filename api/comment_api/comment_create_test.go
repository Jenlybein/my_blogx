package comment_api

import (
	"encoding/json"
	"myblogx/global"
	"myblogx/models"
	"myblogx/service/redis_service/redis_article"
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

func readBizCode(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v body=%s", err, w.Body.String())
	}
	return int(body["code"].(float64))
}

func setupCommentEnv(t *testing.T) *models.UserModel {
	t.Helper()
	_ = testutil.SetupMiniRedis(t)
	db := testutil.SetupSQLite(t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ArticleModel{},
		&models.CommentModel{},
	)
	user := &models.UserModel{
		Username: "comment_u",
		Password: "x",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	return user
}

func TestCommentCreateView(t *testing.T) {
	user := setupCommentEnv(t)
	api := CommentApi{}
	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Username: user.Username}}

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

	t.Run("一级评论成功并写入缓存", func(t *testing.T) {
		c, w := newCommentCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CommentCreateRequest{ArticleID: openArticle.ID, Content: "first"})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code != 0 {
			t.Fatalf("一级评论应成功, body=%s", w.Body.String())
		}

		var cm models.CommentModel
		if err := global.DB.Last(&cm).Error; err != nil {
			t.Fatalf("查询评论失败: %v", err)
		}
		if cm.ParentID != nil || cm.RootParentID != nil {
			t.Fatalf("一级评论 parent/root 应为空: %+v", cm)
		}
		if redis_article.GetCacheComment(openArticle.ID) != 1 {
			t.Fatalf("评论缓存计数错误: %d", redis_article.GetCacheComment(openArticle.ID))
		}
	})

	var first models.CommentModel
	if err := global.DB.Where("article_id = ? and parent_id is null", openArticle.ID).First(&first).Error; err != nil {
		t.Fatalf("获取一级评论失败: %v", err)
	}

	t.Run("回复一级评论成功", func(t *testing.T) {
		c, w := newCommentCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CommentCreateRequest{
			ArticleID: openArticle.ID,
			Content:   "reply",
			ParentID:  &first.ID,
		})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code != 0 {
			t.Fatalf("回复评论应成功, body=%s", w.Body.String())
		}

		var reply models.CommentModel
		if err := global.DB.Last(&reply).Error; err != nil {
			t.Fatalf("查询回复评论失败: %v", err)
		}
		if reply.ParentID == nil || *reply.ParentID != first.ID {
			t.Fatalf("回复评论 parent_id 错误: %+v", reply)
		}
		if reply.RootParentID == nil || *reply.RootParentID != first.ID {
			t.Fatalf("回复评论 root_parent_id 错误: %+v", reply)
		}
	})

	t.Run("父评论不存在", func(t *testing.T) {
		missing := uint(123456)
		c, w := newCommentCtx()
		c.Set("claims", claims)
		c.Set("requestJson", CommentCreateRequest{
			ArticleID: openArticle.ID,
			Content:   "bad",
			ParentID:  &missing,
		})
		api.CommentCreateView(c)
		if code := readBizCode(t, w); code == 0 {
			t.Fatalf("父评论不存在应失败, body=%s", w.Body.String())
		}
	})
}

func TestFindRootParentID(t *testing.T) {
	user := setupCommentEnv(t)
	article := models.ArticleModel{
		Title:          "a",
		Content:        "c",
		AuthorID:       user.ID,
		CommentsToggle: true,
	}
	if err := global.DB.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	root := models.CommentModel{
		Content:   "root",
		UserID:    user.ID,
		ArticleID: article.ID,
	}
	if err := global.DB.Create(&root).Error; err != nil {
		t.Fatalf("创建根评论失败: %v", err)
	}

	child := models.CommentModel{
		Content:      "child",
		UserID:       user.ID,
		ArticleID:    article.ID,
		ParentID:     &root.ID,
		RootParentID: &root.ID,
	}
	if err := global.DB.Create(&child).Error; err != nil {
		t.Fatalf("创建子评论失败: %v", err)
	}

	got, err := findRootParentID(article.ID, child.ID)
	if err != nil {
		t.Fatalf("findRootParentID 失败: %v", err)
	}
	if got == nil || *got != root.ID {
		t.Fatalf("根评论 id 错误: got=%v want=%d", got, root.ID)
	}
}
