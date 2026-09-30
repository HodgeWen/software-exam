package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/middleware"
	"software-exam/backend/internal/service"
)

type ExamHandler struct {
	exams *service.ExamService
}

func NewExamHandler(exams *service.ExamService) *ExamHandler {
	return &ExamHandler{exams: exams}
}

type startExamRequest struct {
	PaperID uint `json:"paper_id" binding:"required"`
}

type submitExamAnswer struct {
	QuestionID uint     `json:"question_id" binding:"required"`
	Selected   []string `json:"selected"`
}

type submitExamRequest struct {
	Answers []submitExamAnswer `json:"answers" binding:"required"`
}

// examResultView 交卷逐题判分视图：对错、所选与正确答案/解析
type examResultView struct {
	QuestionID uint     `json:"question_id"`
	No         int      `json:"no"`
	Selected   []string `json:"selected"`
	Correct    bool     `json:"correct"`
	Answer     []string `json:"answer"`
	Analysis   string   `json:"analysis"`
}

// Start POST /api/v1/exams（鉴权）：开始考试，为当前用户生成进行中的考试记录；
// 返回整卷题目（不含正确答案与解析，与试卷详情共用视图）与试卷时长（分钟）
func (h *ExamHandler) Start(c *gin.Context) {
	var req startExamRequest
	if !bindJSON(c, &req) {
		return
	}
	exam, paper, items, err := h.exams.Start(c.GetUint(middleware.CtxUserID), req.PaperID)
	if errors.Is(err, service.ErrPaperNotFound) {
		Error(c, http.StatusNotFound, "PAPER_NOT_FOUND", "试卷不存在")
		return
	}
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"exam":      exam,
		"paper":     paper,
		"questions": toPaperQuestionViews(items),
	})
}

// Submit POST /api/v1/exams/:id/submit（鉴权）：一次性提交全部作答并统一评分；
// 返回考试记录得分快照（答对题数/总题数/正确率）与逐题对错、正确答案/解析
func (h *ExamHandler) Submit(c *gin.Context) {
	examID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req submitExamRequest
	if !bindJSON(c, &req) {
		return
	}
	// 逐题所选按题目 ID 索引；未出现在提交里的试卷题视为未作答，判错
	selectedByQuestion := make(map[uint][]string, len(req.Answers))
	for _, a := range req.Answers {
		selectedByQuestion[a.QuestionID] = a.Selected
	}
	result, err := h.exams.Submit(c.GetUint(middleware.CtxUserID), examID, selectedByQuestion)
	switch {
	case errors.Is(err, service.ErrExamNotFound):
		Error(c, http.StatusNotFound, "EXAM_NOT_FOUND", "考试记录不存在")
	case errors.Is(err, service.ErrExamAlreadySubmitted):
		Error(c, http.StatusConflict, "EXAM_ALREADY_SUBMITTED", "该考试已交卷")
	case err != nil:
		internalError(c, err)
	default:
		results := make([]examResultView, 0, len(result.Results))
		for _, r := range result.Results {
			results = append(results, examResultView{
				QuestionID: r.Question.ID,
				No:         r.No,
				Selected:   r.Selected,
				Correct:    r.Correct,
				Answer:     r.Question.Answer,
				Analysis:   r.Question.Analysis,
			})
		}
		c.JSON(http.StatusOK, gin.H{"exam": result.Exam, "results": results})
	}
}
