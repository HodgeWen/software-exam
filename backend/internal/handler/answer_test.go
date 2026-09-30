package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/gorm"

	"software-exam/backend/internal/model"
)

// newAnswerTestServer 在通用测试服务上补齐判分与错题相关表，走完整 HTTP 栈
func newAnswerTestServer(t *testing.T) (*httptest.Server, *gorm.DB) {
	t.Helper()
	ts, db := newTestServer(t)
	if err := db.AutoMigrate(
		&model.Subject{}, &model.Chapter{}, &model.Question{},
		&model.AnswerRecord{}, &model.Mistake{}, &model.Favorite{},
	); err != nil {
		t.Fatalf("迁移判分相关表: %v", err)
	}
	return ts, db
}

// userToken 注册并登录指定用户，返回 Authorization 头值
func userToken(t *testing.T, client *http.Client, username string) string {
	t.Helper()
	cred := fmt.Sprintf(`{"username":%q,"password":"pass123"}`, username)
	status, body := doJSON(t, client, http.MethodPost, itBaseURL+"/api/v1/auth/register", cred, "")
	if status != http.StatusCreated {
		t.Fatalf("注册 %s status = %d, body %v", username, status, body)
	}
	status, body = doJSON(t, client, http.MethodPost, itBaseURL+"/api/v1/auth/login", cred, "")
	if status != http.StatusOK {
		t.Fatalf("登录 %s status = %d, body %v", username, status, body)
	}
	return "Bearer " + body["token"].(string)
}

// answerFixtures 两科目三题：A 科单选 + 多选（分属两章节），B 科单选无章节（真题题形态），
// 供科目过滤、分页与统计分布断言
type answerFixtures struct {
	subjA, subjB         model.Subject
	chA1, chA2           model.Chapter
	single, multi, other model.Question
}

var p4Options = []model.Option{
	{Key: "A", Text: "选项A"}, {Key: "B", Text: "选项B"},
	{Key: "C", Text: "选项C"}, {Key: "D", Text: "选项D"},
}

func seedAnswerFixtures(t *testing.T, db *gorm.DB) answerFixtures {
	t.Helper()
	f := answerFixtures{
		subjA: model.Subject{Code: "p4-subj-a", Name: "科目A"},
		subjB: model.Subject{Code: "p4-subj-b", Name: "科目B"},
	}
	if err := db.Create(&f.subjA).Error; err != nil {
		t.Fatalf("种子科目A: %v", err)
	}
	if err := db.Create(&f.subjB).Error; err != nil {
		t.Fatalf("种子科目B: %v", err)
	}
	f.chA1 = model.Chapter{SubjectID: f.subjA.ID, Code: "P4-CH01", Name: "章节一", Sort: 1}
	f.chA2 = model.Chapter{SubjectID: f.subjA.ID, Code: "P4-CH02", Name: "章节二", Sort: 2}
	for _, ch := range []*model.Chapter{&f.chA1, &f.chA2} {
		if err := db.Create(ch).Error; err != nil {
			t.Fatalf("种子章节 %s: %v", ch.Code, err)
		}
	}
	f.single = model.Question{
		SubjectID: f.subjA.ID, ChapterID: &f.chA1.ID, Code: "P4-S01", No: 1, Type: model.QuestionTypeSingle,
		Stem: "A 科单选题干", Options: p4Options, Answer: []string{"A"}, Analysis: "单选解析",
	}
	f.multi = model.Question{
		SubjectID: f.subjA.ID, ChapterID: &f.chA2.ID, Code: "P4-M01", No: 2, Type: model.QuestionTypeMultiple,
		Stem: "A 科多选题干", Options: p4Options, Answer: []string{"A", "C"}, Analysis: "多选解析",
	}
	f.other = model.Question{
		SubjectID: f.subjB.ID, Code: "P4-S02", No: 1, Type: model.QuestionTypeSingle,
		Stem: "B 科单选题干", Options: p4Options, Answer: []string{"B"}, Analysis: "B 科解析",
	}
	for _, q := range []*model.Question{&f.single, &f.multi, &f.other} {
		if err := db.Create(q).Error; err != nil {
			t.Fatalf("种子题目 %s: %v", q.Code, err)
		}
	}
	return f
}

func submitBody(questionID uint, selected string) string {
	return fmt.Sprintf(`{"question_id":%d,"selected":[%s]}`, questionID, selected)
}

// TestSubmitAnswer 单题提交接口：鉴权、参数校验、判分即时返回、记录沉淀
func TestSubmitAnswer(t *testing.T) {
	ts, db := newAnswerTestServer(t)
	client := ts.Client()
	f := seedAnswerFixtures(t, db)
	alice := userToken(t, client, "alice")
	answersURL := itBaseURL + "/api/v1/answers"

	// 未登录 401
	if status, body := doJSON(t, client, http.MethodPost, answersURL,
		submitBody(f.single.ID, `"A"`), ""); status != http.StatusUnauthorized || body["code"] != "UNAUTHORIZED" {
		t.Fatalf("未登录提交 status = %d, body %v", status, body)
	}
	// 缺少所选选项 400
	if status, body := doJSON(t, client, http.MethodPost, answersURL,
		fmt.Sprintf(`{"question_id":%d}`, f.single.ID), alice); status != http.StatusBadRequest || body["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("缺所选选项 status = %d, body %v", status, body)
	}
	// 题目不存在 404
	if status, body := doJSON(t, client, http.MethodPost, answersURL,
		submitBody(99999, `"A"`), alice); status != http.StatusNotFound || body["code"] != "QUESTION_NOT_FOUND" {
		t.Fatalf("题目不存在 status = %d, body %v", status, body)
	}

	// 单选答对：即时返回对/错、正确答案与解析
	status, body := doJSON(t, client, http.MethodPost, answersURL, submitBody(f.single.ID, `"A"`), alice)
	if status != http.StatusOK || body["correct"] != true || body["analysis"] != "单选解析" {
		t.Fatalf("单选答对 status = %d, body %v", status, body)
	}
	if answer, _ := body["answer"].([]any); len(answer) != 1 || answer[0] != "A" {
		t.Fatalf("单选答对应返回正确答案: %v", body["answer"])
	}

	// 多选少选判错：返回正确答案集合
	status, body = doJSON(t, client, http.MethodPost, answersURL, submitBody(f.multi.ID, `"A"`), alice)
	if status != http.StatusOK || body["correct"] != false {
		t.Fatalf("多选少选 status = %d, body %v", status, body)
	}
	if answer, _ := body["answer"].([]any); len(answer) != 2 {
		t.Fatalf("判错应返回正确答案集合: %v", body["answer"])
	}

	// 多选顺序无关判对
	status, body = doJSON(t, client, http.MethodPost, answersURL, submitBody(f.multi.ID, `"C","A"`), alice)
	if status != http.StatusOK || body["correct"] != true {
		t.Fatalf("多选顺序无关 status = %d, body %v", status, body)
	}

	// 每次提交落一条答题记录（3 次提交：2 对 1 错）
	var total, correct int64
	if err := db.Model(&model.AnswerRecord{}).Count(&total).Error; err != nil {
		t.Fatalf("统计答题记录: %v", err)
	}
	if err := db.Model(&model.AnswerRecord{}).Where("correct").Count(&correct).Error; err != nil {
		t.Fatalf("统计答对记录: %v", err)
	}
	if total != 3 || correct != 2 {
		t.Fatalf("答题记录 total=%d correct=%d, want 3/2", total, correct)
	}
}
