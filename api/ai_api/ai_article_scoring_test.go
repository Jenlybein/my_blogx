package ai_api_test

import (
	"encoding/json"
	"myblogx/api/ai_api"
	"myblogx/conf"
	"myblogx/global"
	"myblogx/models/enum"
	"myblogx/service/ai_service"
	"myblogx/utils/jwts"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAIArticleScoringView(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ai_service.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("解析 AI 请求失败: %v", err)
		}
		if len(req.Messages) == 0 {
			t.Fatalf("AI 请求消息不能为空")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role": "assistant",
						"content": `{
							"ai_total_score":88,
							"total_score":88,
							"score_level":"优质文章",
							"article_type":"说明文",
							"dimensions":[
								{"name":"clarity","score":84,"reason":"表达比较清楚"},
								{"name":"structure","score":83,"reason":"结构较顺"},
								{"name":"completeness","score":82,"reason":"内容完整度较好"},
								{"name":"readability","score":80,"reason":"整体较顺畅"},
								{"name":"persuasiveness","score":79,"reason":"论证还有提升空间"},
								{"name":"language","score":85,"reason":"语言较规范"}
							],
							"main_issues":[
								{"positions":[{"paragraph":2,"quote":"这部分论证稍显单薄"}],"reason":"论证支撑不足","suggestion":"补充案例或数据"}
							],
							"overall_comment":"文章整体不错，优先补强论证部分，再顺手压缩重复表述并收束结尾。"
						}`,
					},
				},
			},
		})
	}))
	defer server.Close()

	global.Config = &conf.Config{
		AI: conf.AI{
			Enable:    true,
			SecretKey: "test-key",
			BaseURL:   server.URL,
			ChatModel: "test-model",
		},
	}

	api := ai_api.AIApi{}
	c, w := newAICtx()
	c.Set("claims", &jwts.MyClaims{
		Claims: jwts.Claims{
			UserID:   1,
			Role:     enum.RoleUser,
			Username: "writer",
		},
	})
	c.Set("requestJson", ai_api.AIArticleScoringRequest{
		Title:   "文章标题",
		Content: "# 文章标题\n\n这是一篇测试文章。\n\n它包含完整的正文内容。",
	})

	api.AIArticleScoringView(c)

	if code := readAICode(t, w); code != 0 {
		t.Fatalf("文章评分接口应成功, body=%s", w.Body.String())
	}

	var body struct {
		Code int                             `json:"code"`
		Data ai_api.AIArticleScoringResponse `json:"data"`
		Msg  string                          `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v body=%s", err, w.Body.String())
	}

	if body.Data.TotalScore <= 0 || len(body.Data.Dimensions) != 6 {
		t.Fatalf("评分返回结构错误: %+v", body.Data)
	}
	if body.Data.AITotalScore != 88 || body.Data.ArticleType != "说明文" {
		t.Fatalf("AI 原始总分或文章类型返回错误: %+v", body.Data)
	}
	if len(body.Data.MainIssues) != 1 {
		t.Fatalf("主要问题返回错误: %+v", body.Data.MainIssues)
	}
}
