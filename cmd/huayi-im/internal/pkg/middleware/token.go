package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/response"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/pkg/errno"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/service"
)

// CorsMiddleware 跨域中间件
func TokenMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头获取token
		token := c.GetHeader("Authorization")
		if token == "" {
			response.AbortError(c, errno.Unauthorized.WithMsg("missing token"))
			return
		}

		// 2. 简单处理：Bearer token
		sessionID := token[7:]
		if sessionID == "" {
			response.AbortError(c, errno.Unauthorized.WithMsg("invalid token"))
			return
		}

		// 3. 验证token并获取用户名
		username, exists := userService.GetUsernameBySessionID(sessionID)
		if !exists {
			response.AbortError(c, errno.Unauthorized.WithMsg("invalid token"))
			return
		}

		// 4. 设置用户名到上下文
		c.Set("username", username)
		c.Next()
	}
}
