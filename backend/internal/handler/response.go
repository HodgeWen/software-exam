package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/repository"
)

// CodeInvalidArgument 请求参数校验失败的通用业务错误码
const CodeInvalidArgument = "INVALID_ARGUMENT"

// Error 统一错误响应：{"code": "<业务错误码>", "message": "<人读信息>"}
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"code": code, "message": message})
}

// bindJSON 绑定并校验 JSON 请求体，失败时已写好 400 响应，返回 false 表示终止
func bindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		Error(c, http.StatusBadRequest, CodeInvalidArgument, "请求参数错误: "+err.Error())
		return false
	}
	return true
}

func internalError(c *gin.Context, err error) {
	slog.Error("internal error", "error", err, "path", c.Request.URL.Path)
	Error(c, http.StatusInternalServerError, "INTERNAL", "服务器内部错误")
}

// questionView 题目对外视图：不含正确答案与解析——判分与解析统一由提交接口返回，
// 练习/考试取题接口共用，防止答案经列表接口外泄
type questionView struct {
	ID      uint           `json:"id"`
	No      int            `json:"no"`
	Type    string         `json:"type"`
	Stem    string         `json:"stem"`
	Options []model.Option `json:"options"`
}

func toQuestionView(q model.Question) questionView {
	return questionView{
		ID:      q.ID,
		No:      q.No,
		Type:    q.Type,
		Stem:    q.Stem,
		Options: q.Options,
	}
}

func toQuestionViews(questions []model.Question) []questionView {
	views := make([]questionView, 0, len(questions))
	for _, q := range questions {
		views = append(views, toQuestionView(q))
	}
	return views
}

// questionDetailView 背题模式题目视图：在 questionView 基础上附正确答案与解析，
// 仅供 reveal=1 的取题接口使用；练习/考试取题仍走无答案视图
type questionDetailView struct {
	questionView
	Answer   []string `json:"answer"`
	Analysis string   `json:"analysis"`
}

func toQuestionDetailViews(questions []model.Question) []questionDetailView {
	views := make([]questionDetailView, 0, len(questions))
	for _, q := range questions {
		views = append(views, questionDetailView{
			questionView: toQuestionView(q),
			Answer:       q.Answer,
			Analysis:     q.Analysis,
		})
	}
	return views
}

func toPaperQuestionDetailViews(items []repository.PaperQuestionItem) []questionDetailView {
	views := make([]questionDetailView, 0, len(items))
	for _, item := range items {
		views = append(views, questionDetailView{
			questionView: questionView{
				ID:      item.Question.ID,
				No:      item.No,
				Type:    item.Question.Type,
				Stem:    item.Question.Stem,
				Options: item.Question.Options,
			},
			Answer:   item.Question.Answer,
			Analysis: item.Question.Analysis,
		})
	}
	return views
}

// toPaperQuestionViews 试卷题目项转视图：No 取卷内题号；试卷详情与开始考试共用
func toPaperQuestionViews(items []repository.PaperQuestionItem) []questionView {
	views := make([]questionView, 0, len(items))
	for _, item := range items {
		views = append(views, questionView{
			ID:      item.Question.ID,
			No:      item.No,
			Type:    item.Question.Type,
			Stem:    item.Question.Stem,
			Options: item.Question.Options,
		})
	}
	return views
}

// parseUintParam 解析路径参数为 uint，失败时已写好 400 响应；ID 自 1 起，0 视为不合法
func parseUintParam(c *gin.Context, name string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || v == 0 {
		Error(c, http.StatusBadRequest, CodeInvalidArgument, "路径参数 "+name+" 不合法")
		return 0, false
	}
	return uint(v), true
}

// parseUintQuery 解析可选查询参数为 *uint：缺省返回 nil，给值但不合法时写 400
func parseUintQuery(c *gin.Context, name string) (*uint, bool) {
	raw := c.Query(name)
	if raw == "" {
		return nil, true
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || v == 0 {
		Error(c, http.StatusBadRequest, CodeInvalidArgument, "查询参数 "+name+" 不合法")
		return nil, false
	}
	id := uint(v)
	return &id, true
}
