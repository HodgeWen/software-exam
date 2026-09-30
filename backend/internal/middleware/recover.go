package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recover panic 转统一 500 错误体并记日志，避免裸响应破坏错误契约
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic", "error", err, "stack", string(debug.Stack()))
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{"code": "INTERNAL", "message": "服务器内部错误"})
			}
		}()
		c.Next()
	}
}
