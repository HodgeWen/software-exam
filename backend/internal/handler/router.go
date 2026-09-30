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
	answerService := service.NewAnswerService(
		repository.NewQuestionRepository(db), repository.NewAnswerRecordRepository(db), mistakeRepo)
	answers := NewAnswerHandler(answerService)
	exams := NewExamHandler(service.NewExamService(
		repository.NewBankRepository(db), repository.NewExamRepository(db), answerService))
	mistakes := NewMistakeHandler(service.NewMistakeService(mistakeRepo))
	favorites := NewFavoriteHandler(service.NewFavoriteService(repository.NewFavoriteRepository(db)))
	stats := NewStatsHandler(service.NewStatsService(repository.NewAnswerRecordRepository(db)))

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

	// 真题模拟考试：开始与交卷
	protected.POST("/exams", exams.Start)
	protected.POST("/exams/:id/submit", exams.Submit)

	// 收藏与刷题统计
	protected.POST("/favorites", favorites.Add)
	protected.GET("/favorites", favorites.List)
	protected.DELETE("/favorites/:questionId", favorites.Remove)
	protected.GET("/stats", stats.Summary)

	return r
}
