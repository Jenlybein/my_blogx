package ai_api

import (
	"myblogx/common/res"
	"myblogx/middleware"
	"myblogx/service/ai_service/ai_diagnose"

	"github.com/gin-gonic/gin"
)

// AIDiagnoseView 对选中文本做结构化问题诊断，并通过 SSE 返回最终结构化结果。
func (AIApi) AIDiagnoseView(c *gin.Context) {
	cr := middleware.GetBindJson[AIDiagnoseRequest](c)
	prepareAISSE(c)

	data, err := ai_diagnose.DiagnoseSelectedText(cr.toServiceRequest())
	if err != nil {
		res.SSEFail(err.Error(), c)
		return
	}

	res.SSEOk(data, c)
}
