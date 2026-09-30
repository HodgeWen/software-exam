package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/middleware"
	"software-exam/backend/internal/service"
)

type FavoriteHandler struct {
	favorites *service.FavoriteService
}

func NewFavoriteHandler(favorites *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favorites: favorites}
}

type addFavoriteRequest struct {
	QuestionID uint `json:"question_id" binding:"required"`
}

type listFavoritesQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}

// favoriteItem 收藏列表行：题目走对外视图，正确答案与解析经提交判分接口返回
type favoriteItem struct {
	Question  questionView `json:"question"`
	CreatedAt time.Time    `json:"created_at"`
}

// Add POST /api/v1/favorites（鉴权）：收藏题目；重复收藏不产生重复记录，仍返回成功
func (h *FavoriteHandler) Add(c *gin.Context) {
	var req addFavoriteRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.favorites.Add(c.GetUint(middleware.CtxUserID), req.QuestionID); err != nil {
		internalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// List GET /api/v1/favorites（鉴权）：收藏分页列表（响应含 total），按用户隔离
func (h *FavoriteHandler) List(c *gin.Context) {
	var q listFavoritesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		Error(c, http.StatusBadRequest, CodeInvalidArgument, "请求参数错误: "+err.Error())
		return
	}
	rows, total, err := h.favorites.List(c.GetUint(middleware.CtxUserID), q.Page, q.PageSize)
	if err != nil {
		internalError(c, err)
		return
	}
	items := make([]favoriteItem, len(rows))
	for i, row := range rows {
		items[i] = favoriteItem{Question: toQuestionView(row.Question), CreatedAt: row.Favorite.CreatedAt}
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "list": items})
}

// Remove DELETE /api/v1/favorites/:questionId（鉴权）：取消收藏；未收藏仍返回成功
func (h *FavoriteHandler) Remove(c *gin.Context) {
	questionID, ok := parseUintParam(c, "questionId")
	if !ok {
		return
	}
	if err := h.favorites.Remove(c.GetUint(middleware.CtxUserID), questionID); err != nil {
		internalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
