package ai_diagnose

import "testing"

func TestNormalizeDiagnoseResponse(t *testing.T) {
	resp := normalizeDiagnoseResponse(&DiagnoseResponse{
		Summary: "",
		Issues: []DiagnoseIssue{
			{Type: "未知类型", Severity: "未知", Reason: "句子太长", Evidence: "一句话塞了太多内容", Suggestion: "拆成两句"},
		},
	})

	if resp.Summary == "" {
		t.Fatalf("summary 应有默认值: %+v", resp)
	}
	if len(resp.Issues) != 1 {
		t.Fatalf("issues 数量错误: %+v", resp)
	}
	if resp.Issues[0].Type != "可读性" || resp.Issues[0].Severity != "中" {
		t.Fatalf("类型或严重度规范化失败: %+v", resp.Issues[0])
	}
}

func TestNormalizeDiagnoseRequestTooShort(t *testing.T) {
	_, err := normalizeDiagnoseRequest(DiagnoseRequest{
		SelectionText: "太短了",
		ArticleTitle:  "测试标题",
	})
	if err == nil || err.Error() != selectionTooShortMsg {
		t.Fatalf("短内容应返回固定提示: %v", err)
	}
}
