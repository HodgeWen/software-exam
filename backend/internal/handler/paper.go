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

// PaperDetail GET /api/v1/papers/:id?reveal=（鉴权）：整卷题目按卷内题号升序；
// 默认响应不含正确答案与解析字段（供考试使用），题目视图与练习取题共用；
// reveal=1 时走背题模式视图，附带正确答案与解析
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
	if c.Query("reveal") == "1" {
		c.JSON(http.StatusOK, gin.H{"paper": paper, "questions": toPaperQuestionDetailViews(items)})
		return
	}
	questions := toPaperQuestionViews(items)
	c.JSON(http.StatusOK, gin.H{"paper": paper, "questions": questions})
}
