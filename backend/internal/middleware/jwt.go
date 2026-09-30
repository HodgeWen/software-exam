package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/service"
)

// 鉴权通过后写入 gin.Context 的键
const (
	CtxUserID   = "uid"
	CtxUsername = "username"
)

// JWTAuth 校验 Authorization: Bearer <token>，通过后把用户信息写入上下文；
// 缺失/无效/过期一律 401 统一错误体
func JWTAuth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			unauthorized(c, "缺少 Bearer 令牌")
			return
		}
		claims, err := service.ParseToken(secret, token)
		if err != nil {
			unauthorized(c, "无效或过期的令牌")
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUsername, claims.Username)
		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}

func unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": message})
}
