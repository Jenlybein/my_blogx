package middleware

import (
	"myblogx/common/res"

	"github.com/gin-gonic/gin"
)

func BindJsonMiddleware[T any](c *gin.Context) {
	var cr T
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}
	c.Set("requestJson", cr)
}

func BindQueryMiddleware[T any](c *gin.Context) {
	var cr T
	err := c.ShouldBindQuery(&cr)
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}
	c.Set("requestQuery", cr)
}

func BindUriMiddleware[T any](c *gin.Context) {
	var cr T
	err := c.ShouldBindUri(&cr)
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}
	c.Set("requestUri", cr)
}

func GetBindJson[T any](c *gin.Context) T {
	return c.MustGet("requestJson").(T)
}

func GetBindQuery[T any](c *gin.Context) T {
	return c.MustGet("requestQuery").(T)
}

func GetBindUri[T any](c *gin.Context) T {
	return c.MustGet("requestUri").(T)
}
