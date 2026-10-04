package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/service"
)

// ChapterQuestions GET /api/v1/subjects/:id/chapters/:chapterId/questions?reveal=（鉴权）
// 顺序取题：按题目稳定顺序（ID）返回该章节全部题目；
// reveal=1 时走背题模式视图，附带正确答案与解析
func (h *BankHandler) ChapterQuestions(c *gin.Context) {
	subjectID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	chapterID, ok := parseUintParam(c, "chapterId")
	if !ok {
		return
	}
	questions, err := h.bank.ChapterQuestions(subjectID, chapterID)
	switch {
	case errors.Is(err, service.ErrSubjectNotFound):
		Error(c, http.StatusNotFound, "SUBJECT_NOT_FOUND", "科目不存在")
	case errors.Is(err, service.ErrChapterNotFound):
		Error(c, http.StatusNotFound, "CHAPTER_NOT_FOUND", "章节不存在")
	case err != nil:
		internalError(c, err)
	default:
		if c.Query("reveal") == "1" {
			c.JSON(http.StatusOK, gin.H{"questions": toQuestionDetailViews(questions)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"questions": toQuestionViews(questions)})
	}
}

// RandomQuestions GET /api/v1/subjects/:id/questions/random?chapter_id=（鉴权）
// 随机取题：按科目（可选限定章节）返回打乱后的题目序列
func (h *BankHandler) RandomQuestions(c *gin.Context) {
	subjectID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	chapterID, ok := parseUintQuery(c, "chapter_id")
	if !ok {
		return
	}
	questions, err := h.bank.RandomQuestions(subjectID, chapterID)
	switch {
	case errors.Is(err, service.ErrSubjectNotFound):
		Error(c, http.StatusNotFound, "SUBJECT_NOT_FOUND", "科目不存在")
	case errors.Is(err, service.ErrChapterNotFound):
		Error(c, http.StatusNotFound, "CHAPTER_NOT_FOUND", "章节不存在")
	case err != nil:
		internalError(c, err)
	default:
		c.JSON(http.StatusOK, gin.H{"questions": toQuestionViews(questions)})
	}
}
