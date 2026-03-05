package comment_api

import (
	"encoding/json"
	"myblogx/common"
	"myblogx/global"
	"myblogx/models"
	"myblogx/models/enum"
	"myblogx/service/redis_service/redis_comment"
	"testing"
)

func TestCommentRootListView(t *testing.T) {
	user := setupCommentEnv(t)
	api := CommentApi{}
	if err := global.DB.Model(user).Updates(map[string]any{
		"nickname": "user1",
		"avatar":   "/a1.png",
	}).Error; err != nil {
		t.Fatalf("更新用户1资料失败: %v", err)
	}

	user2 := &models.UserModel{Username: "root_u", Nickname: "user2", Avatar: "/a2.png", Password: "x", Role: enum.RoleUser}
	if err := global.DB.Create(user2).Error; err != nil {
		t.Fatalf("创建第二个用户失败: %v", err)
	}

	article := models.ArticleModel{
		Title:          "open",
		Content:        "c",
		AuthorID:       user.ID,
		CommentsToggle: true,
	}
	if err := global.DB.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	root1 := models.CommentModel{Content: "root1", UserID: user.ID, ArticleID: article.ID, ReplyCount: 2, Status: enum.CommentStatusPublished}
	if err := global.DB.Create(&root1).Error; err != nil {
		t.Fatalf("创建一级评论1失败: %v", err)
	}
	root2 := models.CommentModel{Content: "root2", UserID: user2.ID, ArticleID: article.ID, Status: enum.CommentStatusPublished}
	if err := global.DB.Create(&root2).Error; err != nil {
		t.Fatalf("创建一级评论2失败: %v", err)
	}
	rootPending := models.CommentModel{Content: "root-pending", UserID: user.ID, ArticleID: article.ID, Status: enum.CommentStatusExamining}
	if err := global.DB.Create(&rootPending).Error; err != nil {
		t.Fatalf("创建待审核一级评论失败: %v", err)
	}

	second := models.CommentModel{Content: "reply", UserID: user.ID, ArticleID: article.ID, ReplyId: root1.ID, RootID: root1.ID, Status: enum.CommentStatusPublished}
	if err := global.DB.Create(&second).Error; err != nil {
		t.Fatalf("创建二级评论失败: %v", err)
	}

	if err := redis_comment.SetCacheReply(root1.ID, 1); err != nil {
		t.Fatalf("写入一级评论回复缓存失败: %v", err)
	}
	if err := redis_comment.SetCacheDigg(root1.ID, 2); err != nil {
		t.Fatalf("写入一级评论点赞缓存失败: %v", err)
	}

	t.Run("分页获取一级评论成功", func(t *testing.T) {
		c, w := newCommentCtx()
		c.Set("requestQuery", CommentRootListRequest{
			PageInfo:  common.PageInfo{Limit: 10, Page: 1},
			ArticleID: article.ID,
		})
		api.CommentRootListView(c)
		if code := readBizCode(t, w); code != 0 {
			t.Fatalf("获取一级评论应成功 body=%s", w.Body.String())
		}

		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}
		data := body["data"].(map[string]any)
		if int(data["count"].(float64)) != 2 {
			t.Fatalf("count 错误: %+v", data)
		}

		list := data["list"].([]any)
		if len(list) != 2 {
			t.Fatalf("list 长度错误: %d", len(list))
		}

		var root1Item map[string]any
		for _, raw := range list {
			item := raw.(map[string]any)
			if item["content"] == "root1" {
				root1Item = item
				break
			}
		}
		if root1Item == nil {
			t.Fatalf("未返回 root1: %+v", list)
		}
		if uint(root1Item["reply_id"].(float64)) != 0 {
			t.Fatalf("返回了非一级评论: %+v", root1Item)
		}
		if int(root1Item["reply_count"].(float64)) != 3 {
			t.Fatalf("一级评论 reply_count 未叠加缓存: %+v", root1Item)
		}
		if int(root1Item["digg_count"].(float64)) != 2 {
			t.Fatalf("一级评论 digg_count 未叠加缓存: %+v", root1Item)
		}
		if root1Item["user_nickname"].(string) == "" {
			t.Fatalf("用户昵称为空: %+v", root1Item)
		}
	})

	t.Run("文章不存在时失败", func(t *testing.T) {
		c, w := newCommentCtx()
		c.Set("requestQuery", CommentRootListRequest{
			PageInfo:  common.PageInfo{Limit: 10, Page: 1},
			ArticleID: 999999,
		})
		api.CommentRootListView(c)
		if code := readBizCode(t, w); code == 0 {
			t.Fatalf("文章不存在应失败 body=%s", w.Body.String())
		}
	})
}
