package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/service"
)

// BankHandler 题库取题端点（科目/章节/随机/试卷），方法按资源分文件：subject/practice/paper
type BankHandler struct {
	bank *service.BankService
}

func NewBankHandler(bank *service.BankService) *BankHandler {
	return &BankHandler{bank: bank}
}

// ListSubjects GET /api/v1/subjects（鉴权）：科目列表
func (h *BankHandler) ListSubjects(c *gin.Context) {
	subjects, err := h.bank.Subjects()
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"subjects": subjects})
}

// SubjectDetail GET /api/v1/subjects/:id（鉴权）：科目详情，含章节列表与该科目试卷列表
func (h *BankHandler) SubjectDetail(c *gin.Context) {
	subjectID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	detail, err := h.bank.SubjectDetail(subjectID)
	if errors.Is(err, service.ErrSubjectNotFound) {
		Error(c, http.StatusNotFound, "SUBJECT_NOT_FOUND", "科目不存在")
		return
	}
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"subject":  detail.Subject,
		"chapters": detail.Chapters,
		"papers":   detail.Papers,
	})
}
