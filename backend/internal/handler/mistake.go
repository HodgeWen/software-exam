package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/middleware"
	"software-exam/backend/internal/service"
)

type MistakeHandler struct {
	mistakes *service.MistakeService
}

func NewMistakeHandler(mistakes *service.MistakeService) *MistakeHandler {
	return &MistakeHandler{mistakes: mistakes}
}

type listMistakesQuery struct {
	SubjectID uint `form:"subject_id"`
	Page      int  `form:"page,default=1"`
	PageSize  int  `form:"page_size,default=20"`
}

// mistakeItem 错题本列表行：题目走对外视图，附科目/知识点归属；正确答案与解析经提交判分接口返回
type mistakeItem struct {
	Question    questionView `json:"question"`
	SubjectName string       `json:"subject_name"`
	ChapterName string       `json:"chapter_name"`
	WrongCount  int          `json:"wrong_count"`
	LastWrongAt time.Time    `json:"last_wrong_at"`
}

// List GET /api/v1/mistakes（鉴权）：错题本列表，支持按科目过滤与分页（响应含 total），按用户隔离
func (h *MistakeHandler) List(c *gin.Context) {
	var q listMistakesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		Error(c, http.StatusBadRequest, CodeInvalidArgument, "请求参数错误: "+err.Error())
		return
	}
	rows, total, err := h.mistakes.List(c.GetUint(middleware.CtxUserID), q.SubjectID, q.Page, q.PageSize)
	if err != nil {
		internalError(c, err)
		return
	}
	items := make([]mistakeItem, len(rows))
	for i, row := range rows {
		items[i] = mistakeItem{
			Question:    toQuestionView(row.Question),
			SubjectName: row.SubjectName,
			ChapterName: row.ChapterName,
			WrongCount:  row.Mistake.WrongCount,
			LastWrongAt: row.Mistake.LastWrongAt,
		}
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "list": items})
}

// Questions GET /api/v1/mistakes/questions（鉴权）：错题重刷题源，以当前用户错题为题源返回题目序列
func (h *MistakeHandler) Questions(c *gin.Context) {
	questions, err := h.mistakes.Questions(c.GetUint(middleware.CtxUserID))
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": toQuestionViews(questions)})
}

// Remove DELETE /api/v1/mistakes/:questionId（鉴权）：手动移除错题；重复移除仍成功
func (h *MistakeHandler) Remove(c *gin.Context) {
	questionID, ok := parseUintParam(c, "questionId")
	if !ok {
		return
	}
	if err := h.mistakes.Remove(c.GetUint(middleware.CtxUserID), questionID); err != nil {
		internalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
