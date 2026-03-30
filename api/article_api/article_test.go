package article_api

import (
	"encoding/json"
	"fmt"
	"myblogx/common"
	"myblogx/conf"
	confsite "myblogx/conf/site"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/ctype"
	"myblogx/models/enum"
	"myblogx/models/enum/message_enum"
	"myblogx/test/testutil"
	"myblogx/utils/jwts"
	"myblogx/utils/markdown"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newCtx() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func readCode(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	return int(body["code"].(float64))
}

func readBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	return body
}

func ptrOf[T any](v T) *T {
	return &v
}

func setupArticleEnv(t *testing.T) *models.UserModel {
	t.Helper()
	_ = testutil.SetupMiniRedis(t)
	db := testutil.SetupSQLite(
		t,
		&models.UserModel{},
		&models.UserConfModel{},
		&models.CategoryModel{},
		&models.ArticleModel{},
		&models.TagModel{},
		&models.ArticleTagModel{},
		&models.ArticleDiggModel{},
		&models.FavoriteModel{},
		&models.UserArticleFavorModel{},
		&models.UserTopArticleModel{},
		&models.UserArticleViewHistoryModel{},
		&models.ImageRefModel{},
		&models.CommentModel{},
		&models.ArticleMessageModel{},
	)
	global.Config = &conf.Config{
		Jwt: conf.Jwt{
			Expire: 1,
			Secret: "article-secret",
			Issuer: "article-test",
		},
		Site: conf.Site{
			SiteInfo: confsite.SiteInfo{Mode: enum.SiteModeCommunity},
			Article:  confsite.Article{SkipExamining: false},
		},
	}

	user := &models.UserModel{
		Username: "u1",
		Password: "x",
		Role:     enum.RoleUser,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	return user
}

func waitArticleMessageCount(t *testing.T, want int) []models.ArticleMessageModel {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var list []models.ArticleMessageModel
		if err := global.DB.Order("id asc").Find(&list).Error; err != nil {
			t.Fatalf("查询消息失败: %v", err)
		}
		if len(list) == want {
			return list
		}
		time.Sleep(20 * time.Millisecond)
	}

	var list []models.ArticleMessageModel
	if err := global.DB.Order("id asc").Find(&list).Error; err != nil {
		t.Fatalf("查询消息失败: %v", err)
	}
	t.Fatalf("等待消息数量超时: got=%d want=%d", len(list), want)
	return nil
}

func TestValidateRequestAndList(t *testing.T) {
	user := setupArticleEnv(t)

	{
		c, _ := newCtx()
		if _, err := validateRequest(ArticleListRequest{Type: 1}, nil, c); err != nil {
			t.Fatalf("Type=1 未传 user_id 不应失败: %v", err)
		}
	}

	otherUser := models.UserModel{
		Username: "u2",
		Password: "x",
		Role:     enum.RoleUser,
	}
	if err := global.DB.Create(&otherUser).Error; err != nil {
		t.Fatalf("创建第二个用户失败: %v", err)
	}

	tag := models.TagModel{Title: "Go", IsEnabled: true}
	if err := global.DB.Create(&tag).Error; err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}

	article := models.ArticleModel{
		Title:    "a1",
		Content:  "hello",
		AuthorID: user.ID,
		Status:   enum.ArticleStatusPublished,
	}
	if err := global.DB.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}
	if err := global.DB.Create(&models.ArticleTagModel{ArticleID: article.ID, TagID: tag.ID}).Error; err != nil {
		t.Fatalf("创建文章标签关系失败: %v", err)
	}

	otherArticle := models.ArticleModel{
		Title:    "a2",
		Content:  "world",
		AuthorID: otherUser.ID,
		Status:   enum.ArticleStatusPublished,
	}
	if err := global.DB.Create(&otherArticle).Error; err != nil {
		t.Fatalf("创建第二篇文章失败: %v", err)
	}
	if err := global.DB.Create(&models.ArticleTagModel{ArticleID: otherArticle.ID, TagID: tag.ID}).Error; err != nil {
		t.Fatalf("创建第二篇文章标签关系失败: %v", err)
	}

	draftArticle := models.ArticleModel{
		Title:    "draft",
		Content:  "hidden",
		AuthorID: otherUser.ID,
		Status:   enum.ArticleStatusExamining,
	}
	if err := global.DB.Create(&draftArticle).Error; err != nil {
		t.Fatalf("创建草稿文章失败: %v", err)
	}

	{
		c, w := newCtx()
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("requestQuery", ArticleListRequest{
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
			Type:     1,
			TagID:    &tag.ID,
			Status:   enum.ArticleStatusPublished,
		})
		ArticleApi{}.ArticleListView(c)
		body := readBody(t, w)
		if code := int(body["code"].(float64)); code != 0 {
			t.Fatalf("查询全部公开文章失败, code=%d body=%s", code, w.Body.String())
		}
		data := body["data"].(map[string]any)
		list := data["list"].([]any)
		if count := int(data["count"].(float64)); count != 2 {
			t.Fatalf("未传 user_id 时应查到两篇公开文章, got=%d", count)
		}
		if len(list) != 2 {
			t.Fatalf("未传 user_id 时返回列表长度错误, got=%d", len(list))
		}
	}

	{
		c, w := newCtx()
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("requestQuery", ArticleListRequest{
			PageInfo: common.PageInfo{Page: 1, Limit: 10},
			Type:     1,
			UserID:   user.ID,
			TagID:    &tag.ID,
			Status:   enum.ArticleStatusPublished,
		})
		ArticleApi{}.ArticleListView(c)
		body := readBody(t, w)
		if code := int(body["code"].(float64)); code != 0 {
			t.Fatalf("按 user_id 查询公开文章失败, code=%d body=%s", code, w.Body.String())
		}
		data := body["data"].(map[string]any)
		list := data["list"].([]any)
		if count := int(data["count"].(float64)); count != 1 {
			t.Fatalf("传 user_id=%d 时应只查到一篇文章, got=%d", user.ID, count)
		}
		if len(list) != 1 {
			t.Fatalf("按 user_id 查询时返回列表长度错误, got=%d", len(list))
		}
		item := list[0].(map[string]any)
		if title := item["title"].(string); title != "a1" {
			t.Fatalf("按 user_id 查询返回了错误文章, got=%s", title)
		}
	}
}

func TestArticleCreateUpdateExamineAndRemove(t *testing.T) {
	user := setupArticleEnv(t)
	db := global.DB

	cat := models.CategoryModel{Title: "go", UserID: user.ID}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}
	tag := models.TagModel{Title: "Golang", IsEnabled: true}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}

	api := ArticleApi{}
	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: enum.RoleUser, Username: user.Username}}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", ArticleCreateRequest{
			Title:          "t1",
			Content:        "content",
			CategoryID:     &cat.ID,
			TagIDs:         []ctype.ID{tag.ID},
			CommentsToggle: true,
			Status:         enum.ArticleStatusExamining,
		})
		api.ArticleCreateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("创建文章失败, code=%d body=%s", code, w.Body.String())
		}
	}

	var created models.ArticleModel
	if err := db.Order("id desc").Preload("Tags").First(&created).Error; err != nil {
		t.Fatalf("查询创建文章失败: %v", err)
	}
	if len(created.Tags) != 1 || created.Tags[0].Title != tag.Title {
		t.Fatalf("创建文章后标签关系未正确写入: %+v", created.Tags)
	}

	var relationCount int64
	if err := db.Model(&models.ArticleTagModel{}).
		Where("article_id = ? AND tag_id = ?", created.ID, tag.ID).
		Count(&relationCount).Error; err != nil {
		t.Fatalf("查询文章标签关系失败: %v", err)
	}
	if relationCount != 1 {
		t.Fatalf("文章标签关系应已创建, count=%d", relationCount)
	}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestUri", models.IDRequest{ID: created.ID})
		c.Set("requestJson", ArticleUpdateRequest{
			Title:          ptrOf("t1-updated"),
			Content:        ptrOf("new content"),
			CategoryID:     &cat.ID,
			TagIDs:         &[]ctype.ID{tag.ID},
			CommentsToggle: ptrOf(false),
		})
		api.ArticleUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("更新文章失败, code=%d body=%s", code, w.Body.String())
		}
	}
	if err := db.Preload("Tags").Take(&created, created.ID).Error; err != nil {
		t.Fatalf("回查更新后的文章失败: %v", err)
	}
	if len(created.Tags) != 1 || created.Tags[0].Title != tag.Title {
		t.Fatalf("更新文章后标签关系未正确写入: %+v", created.Tags)
	}

	{
		c, w := newCtx()
		c.Set("requestUri", models.IDRequest{ID: created.ID})
		c.Set("requestJson", ArticleExamineRequest{Status: enum.ArticleStatusPublished})
		api.ArticleExamineView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("审核文章失败, code=%d body=%s", code, w.Body.String())
		}
	}

	messages := waitArticleMessageCount(t, 1)
	if messages[0].Type != message_enum.SystemType {
		t.Fatalf("文章审核消息类型错误: %+v", messages[0])
	}
	if messages[0].ReceiverID != user.ID {
		t.Fatalf("文章审核消息接收者错误: %+v", messages[0])
	}
	if messages[0].Content != fmt.Sprintf("您的文章《%s》审核通过!", "t1-updated") {
		t.Fatalf("文章审核消息内容错误: %+v", messages[0])
	}
	if messages[0].LinkHerf != fmt.Sprintf("/article/%d", created.ID) {
		t.Fatalf("文章审核消息链接错误: %+v", messages[0])
	}

	{
		c, w := newCtx()
		c.Set("requestJson", models.IDListRequest{IDList: []ctype.ID{created.ID}})
		api.ArticleRemoveView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("删除文章失败, code=%d body=%s", code, w.Body.String())
		}
	}
}

func TestArticleUpdateViewOnlyUpdatesProvidedFields(t *testing.T) {
	user := setupArticleEnv(t)
	db := global.DB

	cat := models.CategoryModel{Title: "go", UserID: user.ID}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}
	tag := models.TagModel{Title: "Golang", IsEnabled: true}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}

	article := models.ArticleModel{
		Title:          "old title",
		Abstract:       "manual abstract",
		Content:        "old content",
		CategoryID:     &cat.ID,
		Cover:          "old-cover",
		AuthorID:       user.ID,
		CommentsToggle: true,
		Status:         enum.ArticleStatusExamining,
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}
	if err := db.Create(&models.ArticleTagModel{ArticleID: article.ID, TagID: tag.ID}).Error; err != nil {
		t.Fatalf("创建文章标签关系失败: %v", err)
	}

	api := ArticleApi{}
	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: enum.RoleUser, Username: user.Username}}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		c.Set("requestJson", ArticleUpdateRequest{
			Content: ptrOf("new content"),
		})
		api.ArticleUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("部分更新文章失败, code=%d body=%s", code, w.Body.String())
		}
	}

	if err := db.Preload("Tags").Take(&article, article.ID).Error; err != nil {
		t.Fatalf("回查更新后的文章失败: %v", err)
	}
	if article.Title != "old title" {
		t.Fatalf("未传 title 不应更新, got=%s", article.Title)
	}
	if article.Abstract != "manual abstract" {
		t.Fatalf("未传 abstract 不应更新, got=%s", article.Abstract)
	}
	if article.Content != markdown.MdToSafe("new content") {
		t.Fatalf("已传 content 应更新为安全 Markdown, got=%q", article.Content)
	}
	if article.Cover != "old-cover" {
		t.Fatalf("未传 cover 不应更新, got=%s", article.Cover)
	}
	if !article.CommentsToggle {
		t.Fatal("未传 comments_toggle 不应更新")
	}
	if article.CategoryID == nil || *article.CategoryID != cat.ID {
		t.Fatalf("未传 category_id 不应更新, got=%v", article.CategoryID)
	}
	if len(article.Tags) != 1 || article.Tags[0].Title != tag.Title {
		t.Fatalf("未传 tag_ids 不应更新标签关系, got=%+v", article.Tags)
	}

	var relationCount int64
	if err := db.Model(&models.ArticleTagModel{}).
		Where("article_id = ? AND tag_id = ?", article.ID, tag.ID).
		Count(&relationCount).Error; err != nil {
		t.Fatalf("查询文章标签关系失败: %v", err)
	}
	if relationCount != 1 {
		t.Fatalf("未传 tag_ids 不应更新标签关系, count=%d", relationCount)
	}
}

func TestArticleUpdateViewCategoryIDZeroClearsCategory(t *testing.T) {
	user := setupArticleEnv(t)
	db := global.DB

	cat := models.CategoryModel{Title: "go", UserID: user.ID}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}

	article := models.ArticleModel{
		Title:      "old title",
		Content:    "old content",
		CategoryID: &cat.ID,
		AuthorID:   user.ID,
		Status:     enum.ArticleStatusExamining,
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	api := ArticleApi{}
	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: enum.RoleUser, Username: user.Username}}
	clearID := ctype.ID(0)

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		c.Set("requestJson", ArticleUpdateRequest{
			CategoryID: &clearID,
		})
		api.ArticleUpdateView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("清空文章分类失败, code=%d body=%s", code, w.Body.String())
		}
	}

	var updated models.ArticleModel
	if err := db.Take(&updated, article.ID).Error; err != nil {
		t.Fatalf("回查清空分类后的文章失败: %v", err)
	}
	if updated.CategoryID != nil {
		t.Fatalf("传 category_id=0 应清空分类, got=%v", *updated.CategoryID)
	}
}

func TestArticleDiggFavoriteVisitDetailRemoveUser(t *testing.T) {
	user := setupArticleEnv(t)
	db := global.DB
	api := ArticleApi{}

	tag := models.TagModel{Title: "Backend", IsEnabled: true}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}

	article := models.ArticleModel{
		Title:    "a1",
		Content:  "content",
		AuthorID: user.ID,
		Status:   enum.ArticleStatusPublished,
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}
	if err := db.Create(&models.ArticleTagModel{ArticleID: article.ID, TagID: tag.ID}).Error; err != nil {
		t.Fatalf("创建文章标签关系失败: %v", err)
	}

	claims := &jwts.MyClaims{Claims: jwts.Claims{UserID: user.ID, Role: enum.RoleUser, Username: user.Username}}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		api.ArticleDiggView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("点赞失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", ArticleFavoriteRequest{ArticleID: article.ID})
		api.ArticleFavoriteSaveView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("收藏失败, code=%d body=%s", code, w.Body.String())
		}
	}

	token := testutil.IssueAccessToken(t, user)
	{
		c, w := newCtx()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("token", token)
		c.Request = req
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		api.ArticleDetailView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("文章详情失败, code=%d body=%s", code, w.Body.String())
		}
		body := readBody(t, w)
		data := body["data"].(map[string]any)
		if !data["is_digg"].(bool) || !data["is_favor"].(bool) {
			t.Fatalf("文章详情点赞/收藏状态异常, body=%s", w.Body.String())
		}
	}
	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestJson", ArticleFavoriteRequest{ArticleID: article.ID})
		api.ArticleFavoriteSaveView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("取消收藏失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
		c.Set("requestJson", ArticleViewCountRequest{ArticleID: article.ID})
		api.ArticleVisitView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("访问计数失败, code=%d body=%s", code, w.Body.String())
		}
	}
	{
		c, w := newCtx()
		c.Set("claims", claims)
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		api.ArticleDiggView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("取消点赞失败, code=%d body=%s", code, w.Body.String())
		}
	}

	{
		c, w := newCtx()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("token", token)
		c.Request = req
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		api.ArticleDetailView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("文章详情失败, code=%d body=%s", code, w.Body.String())
		}
		body := readBody(t, w)
		data := body["data"].(map[string]any)
		if data["is_digg"].(bool) || data["is_favor"].(bool) {
			t.Fatalf("取消后文章详情点赞/收藏状态异常, body=%s", w.Body.String())
		}
	}

	{
		c, w := newCtx()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		req.Header.Set("token", token)
		c.Request = req
		c.Set("requestUri", models.IDRequest{ID: article.ID})
		api.ArticleRemoveUserView(c)
		if code := readCode(t, w); code != 0 {
			t.Fatalf("用户删除文章失败, code=%d body=%s", code, w.Body.String())
		}
	}
}
