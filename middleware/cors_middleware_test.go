package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"myblogx/middleware"

	"github.com/gin-gonic/gin"
)

func TestCorsMiddlewarePreflightAndActualRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CorsMiddleware())
	r.POST("/api/users/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	origin := "http://localhost:5173"

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/api/users/login", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		req.Header.Set("Access-Control-Request-Headers", "token, content-type")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("预检请求状态码异常: %d", w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Fatalf("预检请求未返回允许来源: %q", got)
		}
		if got := strings.ToLower(w.Header().Get("Access-Control-Allow-Headers")); !strings.Contains(got, "token") {
			t.Fatalf("预检请求未放行 token 请求头: %q", got)
		}
	}

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/users/login", nil)
		req.Header.Set("Origin", origin)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("实际请求状态码异常: %d", w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Fatalf("实际请求未返回允许来源: %q", got)
		}
	}
}
