package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/middleware"
	"software-exam/backend/internal/service"
)

type AnswerHandler struct {
	answers *service.AnswerService
}

func NewAnswerHandler(answers *service.AnswerService) *AnswerHandler {
	return &AnswerHandler{answers: answers}
}

type submitAnswerRequest struct {
	QuestionID uint     `json:"question_id" binding:"required"`
	Selected   []string `json:"selected" binding:"required,min=1"`
}

// Submit POST /api/v1/answers（鉴权）：练习与错题重刷共用的单题提交，服务端判分并即时返回对/错、正确答案与解析
func (h *AnswerHandler) Submit(c *gin.Context) {
	var req submitAnswerRequest
	if !bindJSON(c, &req) {
		return
	}
	q, correct, err := h.answers.Submit(c.GetUint(middleware.CtxUserID), req.QuestionID, req.Selected)
	switch {
	case errors.Is(err, service.ErrQuestionNotFound):
		Error(c, http.StatusNotFound, "QUESTION_NOT_FOUND", "题目不存在")
	case err != nil:
		internalError(c, err)
	default:
		c.JSON(http.StatusOK, gin.H{
			"correct":  correct,
			"answer":   q.Answer,
			"analysis": q.Analysis,
		})
	}
}
