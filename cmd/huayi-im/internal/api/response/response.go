package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/pkg/errno"
)

func Success(c *gin.Context, data interface{}) {
	if data == nil {
		// 如果data为nil，则只返回状态码200，不返回body
		c.Status(http.StatusOK)
		return
	}
	c.JSON(http.StatusOK, data)
}

// Error 失败响应
func Error(c *gin.Context, err errno.Errno) {
	c.JSON(err.Code(), gin.H{"error": err.Message()})
}

// AbortError 终止请求并返回失败响应
func AbortError(c *gin.Context, err errno.Errno) {
	c.Abort()
	Error(c, err)
}
