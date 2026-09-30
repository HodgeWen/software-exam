package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"software-exam/backend/internal/middleware"
	"software-exam/backend/internal/repository"
	"software-exam/backend/internal/service"
)

// NewRouter 装配中间件与全部路由；集成测试用同一入口挂内存 SQLite
func NewRouter(db *gorm.DB, jwtSecret []byte) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recover(), middleware.Logger(), middleware.CORS())

	auth := NewAuthHandler(service.NewAuthService(repository.NewUserRepository(db), jwtSecret))

	v1 := r.Group("/api/v1")
	v1.GET("/health", Health)

	// 公开路由：注册/登录
	pub := v1.Group("/auth")
	pub.POST("/register", auth.Register)
	pub.POST("/login", auth.Login)

	// 其余 /api/v1 接口一律 JWT 鉴权，后续阶段的路由挂到该组
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(jwtSecret))
	protected.GET("/me", auth.Me)

	return r
}
