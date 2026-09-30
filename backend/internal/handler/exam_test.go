package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/gorm"

	"software-exam/backend/internal/model"
)

// newExamTestServer 在通用测试服务上补齐考试相关表，走完整 HTTP 栈
func newExamTestServer(t *testing.T) (*httptest.Server, *gorm.DB) {
	t.Helper()
	ts, db := newTestServer(t)
	if err := db.AutoMigrate(
		&model.Subject{}, &model.Question{},
		&model.Paper{}, &model.PaperQuestion{},
		&model.Exam{}, &model.AnswerRecord{}, &model.Mistake{},
	); err != nil {
		t.Fatalf("迁移考试相关表: %v", err)
	}
	return ts, db
}

// examFixtures 一套 4 题试卷：单选×2、多选×2，覆盖答对/少选/多选/未作答四种结局
type examFixtures struct {
	paper                              model.Paper
	singleA, multiAC, multiBD, singleC model.Question
}

func seedExamFixtures(t *testing.T, db *gorm.DB) *examFixtures {
	t.Helper()
	subj := model.Subject{Code: "p6-subj", Name: "考试科目"}
	if err := db.Create(&subj).Error; err != nil {
		t.Fatalf("种子科目: %v", err)
	}
	f := &examFixtures{
		paper: model.Paper{SubjectID: subj.ID, Code: "p6-paper", Name: "模拟卷", DurationMinutes: 60},
	}
	if err := db.Create(&f.paper).Error; err != nil {
		t.Fatalf("种子试卷: %v", err)
	}
	f.singleA = model.Question{SubjectID: subj.ID, Code: "P6-S01", No: 1, Type: model.QuestionTypeSingle,
		Stem: "单选一", Options: p4Options, Answer: []string{"A"}, Analysis: "解析一"}
	f.multiAC = model.Question{SubjectID: subj.ID, Code: "P6-M01", No: 2, Type: model.QuestionTypeMultiple,
		Stem: "多选一", Options: p4Options, Answer: []string{"A", "C"}, Analysis: "解析二"}
	f.multiBD = model.Question{SubjectID: subj.ID, Code: "P6-M02", No: 3, Type: model.QuestionTypeMultiple,
		Stem: "多选二", Options: p4Options, Answer: []string{"B", "D"}, Analysis: "解析三"}
	f.singleC = model.Question{SubjectID: subj.ID, Code: "P6-S02", No: 4, Type: model.QuestionTypeSingle,
		Stem: "单选二", Options: p4Options, Answer: []string{"C"}, Analysis: "解析四"}
	for i, q := range []*model.Question{&f.singleA, &f.multiAC, &f.multiBD, &f.singleC} {
		if err := db.Create(q).Error; err != nil {
			t.Fatalf("种子题目 %s: %v", q.Code, err)
		}
		link := model.PaperQuestion{PaperID: f.paper.ID, QuestionID: q.ID, No: i + 1}
		if err := db.Create(&link).Error; err != nil {
			t.Fatalf("关联试卷题目 %s: %v", q.Code, err)
		}
	}
	return f
}

// submitExamBody 组装交卷请求体：逐题题目 ID 与所选集合
func submitExamBody(items ...[2]any) string {
	body := ""
	for i, it := range items {
		if i > 0 {
			body += ","
		}
		body += fmt.Sprintf(`{"question_id":%d,"selected":[%s]}`, it[0], it[1])
	}
	return `{"answers":[` + body + `]}`
}

// TestExamFlow 模拟考试全流程：开始（整卷题目+时长+不泄答案）、交卷统一评分
// （每题对错/得分/逐题答案解析）、落答题记录与错题本、重复交卷业务错误、用户隔离
func TestExamFlow(t *testing.T) {
	ts, db := newExamTestServer(t)
	client := ts.Client()
	f := seedExamFixtures(t, db)
	alice := userToken(t, client, "exam-alice")
	bob := userToken(t, client, "exam-bob")
	examsURL := itBaseURL + "/api/v1/exams"

	// 未登录 401；缺参 400；未知试卷 404
	if status, body := doJSON(t, client, http.MethodPost, examsURL,
		fmt.Sprintf(`{"paper_id":%d}`, f.paper.ID), ""); status != http.StatusUnauthorized || body["code"] != "UNAUTHORIZED" {
		t.Fatalf("未登录开始考试 status = %d, body %v", status, body)
	}
	if status, body := doJSON(t, client, http.MethodPost, examsURL, `{}`, alice); status != http.StatusBadRequest || body["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("缺 paper_id status = %d, body %v", status, body)
	}
	if status, body := doJSON(t, client, http.MethodPost, examsURL, `{"paper_id":99999}`, alice); status != http.StatusNotFound || body["code"] != "PAPER_NOT_FOUND" {
		t.Fatalf("未知试卷 status = %d, body %v", status, body)
	}

	// 开始考试：进行中考试记录 + 试卷时长 + 整卷题目按题号升序且不含答案与解析
	status, body := doJSON(t, client, http.MethodPost, examsURL,
		fmt.Sprintf(`{"paper_id":%d}`, f.paper.ID), alice)
	if status != http.StatusCreated {
		t.Fatalf("开始考试 status = %d, body %v", status, body)
	}
	exam := body["exam"].(map[string]any)
	if exam["status"] != model.ExamStatusInProgress {
		t.Fatalf("新考试应进行中: %v", exam)
	}
	if body["paper"].(map[string]any)["duration_minutes"] != float64(60) {
		t.Fatalf("应返回试卷时长（分钟）: %v", body["paper"])
	}
	questions := body["questions"].([]any)
	if len(questions) != 4 {
		t.Fatalf("整卷题目数 = %d, want 4", len(questions))
	}
	assertNoAnswerLeak(t, questions)
	for i, item := range questions {
		if q := item.(map[string]any); q["no"].(float64) != float64(i+1) {
			t.Fatalf("整卷题目应按题号升序: %v", questions)
		}
	}
	examID := uint(exam["id"].(float64))

	// 交卷：单选一答对、多选一少选、多选二多选、单选二未提交（未作答判错）
	submitURL := fmt.Sprintf("%s/%d/submit", examsURL, examID)
	payload := submitExamBody(
		[2]any{f.singleA.ID, `"A"`},
		[2]any{f.multiAC.ID, `"A"`},
		[2]any{f.multiBD.ID, `"A","B","D"`},
	)
	status, body = doJSON(t, client, http.MethodPost, submitURL, payload, alice)
	if status != http.StatusOK {
		t.Fatalf("交卷 status = %d, body %v", status, body)
	}
	exam = body["exam"].(map[string]any)
	if exam["status"] != model.ExamStatusSubmitted || exam["submitted_at"] == nil {
		t.Fatalf("交卷后考试应已提交: %v", exam)
	}
	if exam["correct_count"] != float64(1) || exam["total_count"] != float64(4) || exam["accuracy"] != 0.25 {
		t.Fatalf("交卷得分不符: %v", exam)
	}
	results, _ := body["results"].([]any)
	if len(results) != 4 {
		t.Fatalf("逐题结果数 = %d, want 4", len(results))
	}
	wantCorrect := map[float64]bool{1: true, 2: false, 3: false, 4: false}
	for _, item := range results {
		r := item.(map[string]any)
		no := r["no"].(float64)
		if r["correct"] != wantCorrect[no] {
			t.Fatalf("第 %v 题判分 = %v, want %v: %v", no, r["correct"], wantCorrect[no], r)
		}
		if len(r["answer"].([]any)) == 0 || r["analysis"] == "" {
			t.Fatalf("逐题结果应带正确答案与解析: %v", r)
		}
	}
	if r := results[3].(map[string]any); r["selected"] != nil {
		t.Fatalf("未作答题所选应为空: %v", r)
	}

	// 沉淀断言：4 条答题记录（1 对）、3 道错题入错题本、考试记录已交卷
	var records, mistakes int64
	if err := db.Model(&model.AnswerRecord{}).Where("user_id = ?", 1).Count(&records).Error; err != nil || records != 4 {
		t.Fatalf("考试应按题落答题记录 = %d (err %v), want 4", records, err)
	}
	if err := db.Model(&model.Mistake{}).Where("user_id = ?", 1).Count(&mistakes).Error; err != nil || mistakes != 3 {
		t.Fatalf("考试答错应入错题本 = %d (err %v), want 3", mistakes, err)
	}
	var stored model.Exam
	if err := db.First(&stored, examID).Error; err != nil || stored.Status != model.ExamStatusSubmitted ||
		stored.CorrectCount != 1 || stored.TotalCount != 4 || stored.SubmittedAt == nil {
		t.Fatalf("落库考试记录不符: %+v (err %v)", stored, err)
	}

	// 重复交卷：409 业务错误，且不再新增答题记录
	if status, body = doJSON(t, client, http.MethodPost, submitURL, payload, alice); status != http.StatusConflict || body["code"] != "EXAM_ALREADY_SUBMITTED" {
		t.Fatalf("重复交卷 status = %d, body %v", status, body)
	}
	if err := db.Model(&model.AnswerRecord{}).Where("user_id = ?", 1).Count(&records).Error; err != nil || records != 4 {
		t.Fatalf("重复交卷不应新增答题记录 = %d (err %v)", records, err)
	}

	// 用户隔离：他人交卷视为考试不存在；非法与未知 ID 分别 400 / 404
	if status, body = doJSON(t, client, http.MethodPost, submitURL, payload, bob); status != http.StatusNotFound || body["code"] != "EXAM_NOT_FOUND" {
		t.Fatalf("他人交卷 status = %d, body %v", status, body)
	}
	if status, body = doJSON(t, client, http.MethodPost, examsURL+"/abc/submit", payload, alice); status != http.StatusBadRequest || body["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("非法考试 ID status = %d, body %v", status, body)
	}
	if status, body = doJSON(t, client, http.MethodPost, examsURL+"/99999/submit", payload, alice); status != http.StatusNotFound || body["code"] != "EXAM_NOT_FOUND" {
		t.Fatalf("未知考试 status = %d, body %v", status, body)
	}
}
