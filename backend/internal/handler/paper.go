package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/service"
)

// ListPapers GET /api/v1/papers（鉴权）：试卷列表
func (h *BankHandler) ListPapers(c *gin.Context) {
	papers, err := h.bank.Papers()
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"papers": papers})
}

// PaperDetail GET /api/v1/papers/:id（鉴权）：整卷题目按卷内题号升序；
// 响应不含正确答案与解析字段（供考试使用），题目视图与练习取题共用
func (h *BankHandler) PaperDetail(c *gin.Context) {
	paperID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	paper, items, err := h.bank.PaperQuestions(paperID)
	if errors.Is(err, service.ErrPaperNotFound) {
		Error(c, http.StatusNotFound, "PAPER_NOT_FOUND", "试卷不存在")
		return
	}
	if err != nil {
		internalError(c, err)
		return
	}
	questions := make([]questionView, 0, len(items))
	for _, item := range items {
		q := item.Question
		questions = append(questions, questionView{
			ID:      q.ID,
			No:      item.No,
			Type:    q.Type,
			Stem:    q.Stem,
			Options: q.Options,
		})
	}
	c.JSON(http.StatusOK, gin.H{"paper": paper, "questions": questions})
}
