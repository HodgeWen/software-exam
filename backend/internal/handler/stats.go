package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/middleware"
	"software-exam/backend/internal/service"
)

type StatsHandler struct {
	stats *service.StatsService
}

func NewStatsHandler(stats *service.StatsService) *StatsHandler {
	return &StatsHandler{stats: stats}
}

// Summary GET /api/v1/stats（鉴权）：当前用户的总答题量、总正确率与按科目/章节的分布
func (h *StatsHandler) Summary(c *gin.Context) {
	summary, err := h.stats.Summary(c.GetUint(middleware.CtxUserID))
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}
