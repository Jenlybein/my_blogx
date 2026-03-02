package service_test

import (
	"myblogx/service/redis_service/redis_article"
	"myblogx/test/testutil"
	"testing"
)

func TestArticleCacheCounters(t *testing.T) {
	_ = testutil.SetupMiniRedis(t)

	if err := redis_article.SetCacheView(1, 3); err != nil {
		t.Fatalf("SetCacheView 失败: %v", err)
	}
	if err := redis_article.SetCacheDigg(1, 2); err != nil {
		t.Fatalf("SetCacheDigg 失败: %v", err)
	}
	if err := redis_article.SetCacheFavorite(1, 1); err != nil {
		t.Fatalf("SetCacheFavorite 失败: %v", err)
	}
	if err := redis_article.SetCacheComment(1, 4); err != nil {
		t.Fatalf("SetCacheComment 失败: %v", err)
	}

	if redis_article.GetCacheView(1) != 3 {
		t.Fatalf("view 计数错误: %d", redis_article.GetCacheView(1))
	}
	if redis_article.GetCacheDigg(1) != 2 {
		t.Fatalf("digg 计数错误: %d", redis_article.GetCacheDigg(1))
	}
	if redis_article.GetCacheFavorite(1) != 1 {
		t.Fatalf("favorite 计数错误: %d", redis_article.GetCacheFavorite(1))
	}
	if redis_article.GetCacheComment(1) != 4 {
		t.Fatalf("comment 计数错误: %d", redis_article.GetCacheComment(1))
	}

	if len(redis_article.GetAllCacheView()) != 1 {
		t.Fatal("GetAllCacheView 长度异常")
	}

	if err := redis_article.ClearAllCacheArticle(); err != nil {
		t.Fatalf("ClearAllCacheArticle 失败: %v", err)
	}
	if redis_article.GetCacheView(1) != 0 ||
		redis_article.GetCacheDigg(1) != 0 ||
		redis_article.GetCacheFavorite(1) != 0 ||
		redis_article.GetCacheComment(1) != 0 {
		t.Fatal("清理后计数应为 0")
	}
}

func TestArticleHistoryCache(t *testing.T) {
	_ = testutil.SetupMiniRedis(t)

	redis_article.SetUserArticleHistoryCache(10, 20)
	if !redis_article.GetUserArticleHistoryCache(10, 20) {
		t.Fatal("用户历史记录应存在")
	}

	redis_article.SetGuestArticleHistoryCache(11, "abc")
	if !redis_article.GetGuestArticleHistoryCache(11, "abc") {
		t.Fatal("访客历史记录应存在")
	}
}
