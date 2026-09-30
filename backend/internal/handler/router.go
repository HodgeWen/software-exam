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
	bank := NewBankHandler(service.NewBankService(repository.NewBankRepository(db)))
	mistakeRepo := repository.NewMistakeRepository(db)
	answers := NewAnswerHandler(service.NewAnswerService(
		repository.NewQuestionRepository(db), repository.NewAnswerRecordRepository(db), mistakeRepo))
	mistakes := NewMistakeHandler(service.NewMistakeService(mistakeRepo))

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

	// 题库取题：科目/章节（顺序）/随机/试卷
	protected.GET("/subjects", bank.ListSubjects)
	protected.GET("/subjects/:id", bank.SubjectDetail)
	protected.GET("/subjects/:id/chapters/:chapterId/questions", bank.ChapterQuestions)
	protected.GET("/subjects/:id/questions/random", bank.RandomQuestions)
	protected.GET("/papers", bank.ListPapers)
	protected.GET("/papers/:id", bank.PaperDetail)

	// 刷题判分与错题本
	protected.POST("/answers", answers.Submit)
	protected.GET("/mistakes", mistakes.List)
	protected.GET("/mistakes/questions", mistakes.Questions)
	protected.DELETE("/mistakes/:questionId", mistakes.Remove)

	return r
}
