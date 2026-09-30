package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	"gorm.io/gorm"

	"software-exam/backend/internal/model"
	"software-exam/backend/seed"
)

// newBankTestServer 在通用测试服务上补齐题库表并导入真实种子，走完整 HTTP 栈
func newBankTestServer(t *testing.T) (*httptest.Server, *gorm.DB) {
	t.Helper()
	ts, db := newTestServer(t)
	if err := db.AutoMigrate(
		&model.Subject{}, &model.Chapter{}, &model.Question{},
		&model.Paper{}, &model.PaperQuestion{},
	); err != nil {
		t.Fatalf("迁移题库表: %v", err)
	}
	if err := seed.Import(db); err != nil {
		t.Fatalf("导入种子: %v", err)
	}
	return ts, db
}

// bankToken 注册并登录测试用户，返回 Authorization 头值
func bankToken(t *testing.T, client *http.Client) string {
	t.Helper()
	ts := "http://example.com"
	status, body := doJSON(t, client, http.MethodPost, ts+"/api/v1/auth/register", `{"username":"banker","password":"pass123"}`, "")
	if status != http.StatusCreated {
		t.Fatalf("注册测试用户 status = %d, body %v", status, body)
	}
	status, body = doJSON(t, client, http.MethodPost, ts+"/api/v1/auth/login", `{"username":"banker","password":"pass123"}`, "")
	if status != http.StatusOK {
		t.Fatalf("登录测试用户 status = %d, body %v", status, body)
	}
	return "Bearer " + body["token"].(string)
}

// questionIDsFromResp 从响应 JSON 里取题目 ID 集合（乱序断言用，返回前先排序）
func questionIDsFromResp(t *testing.T, items []any) []float64 {
	t.Helper()
	ids := make([]float64, 0, len(items))
	for _, item := range items {
		q, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("题目项不是对象: %v", item)
		}
		ids = append(ids, q["id"].(float64))
	}
	slices.Sort(ids)
	return ids
}

// assertNoAnswerLeak 断言每个题目对象都不含正确答案与解析字段
func assertNoAnswerLeak(t *testing.T, items []any) {
	t.Helper()
	for _, item := range items {
		q := item.(map[string]any)
		for _, key := range []string{"answer", "analysis"} {
			if _, has := q[key]; has {
				t.Fatalf("题目响应泄露 %s 字段: %v", key, q)
			}
		}
	}
}

func TestBankEndpointsRequireAuth(t *testing.T) {
	ts, _ := newBankTestServer(t)
	client := ts.Client()
	for _, path := range []string{"/api/v1/subjects", "/api/v1/papers"} {
		if status, body := doJSON(t, client, http.MethodGet, itBaseURL+path, "", ""); status != http.StatusUnauthorized || body["code"] != "UNAUTHORIZED" {
			t.Fatalf("无 token 访问 %s status = %d, body %v", path, status, body)
		}
	}
}

func TestSubjectsListAndDetail(t *testing.T) {
	ts, db := newBankTestServer(t)
	auth := bankToken(t, ts.Client())

	var subject model.Subject
	if err := db.Where("code = ?", "soft-designer").First(&subject).Error; err != nil {
		t.Fatalf("查种子科目: %v", err)
	}

	// 科目列表：包含种子科目
	status, body := doJSON(t, ts.Client(), http.MethodGet, itBaseURL+"/api/v1/subjects", "", auth)
	if status != http.StatusOK {
		t.Fatalf("科目列表 status = %d, body %v", status, body)
	}
	found := false
	for _, s := range body["subjects"].([]any) {
		if s.(map[string]any)["code"] == "soft-designer" {
			found = true
		}
	}
	if !found {
		t.Fatalf("科目列表缺少 soft-designer: %v", body)
	}

	// 科目详情：章节按 sort 升序、试卷归属该科目
	var wantChapters []model.Chapter
	if err := db.Where("subject_id = ?", subject.ID).Order("sort").Find(&wantChapters).Error; err != nil {
		t.Fatalf("查章节: %v", err)
	}
	var wantPapers []model.Paper
	if err := db.Where("subject_id = ?", subject.ID).Order("id").Find(&wantPapers).Error; err != nil {
		t.Fatalf("查试卷: %v", err)
	}

	status, body = doJSON(t, ts.Client(), http.MethodGet, itBaseURL+"/api/v1/subjects/"+itoa(subject.ID), "", auth)
	if status != http.StatusOK {
		t.Fatalf("科目详情 status = %d, body %v", status, body)
	}
	chapters := body["chapters"].([]any)
	if len(chapters) != len(wantChapters) {
		t.Fatalf("详情章节数 = %d, want %d", len(chapters), len(wantChapters))
	}
	for i, ch := range chapters {
		got := ch.(map[string]any)
		if got["code"] != wantChapters[i].Code || got["sort"].(float64) != float64(wantChapters[i].Sort) {
			t.Fatalf("第 %d 个章节 = %v, want %v", i, got, wantChapters[i].Code)
		}
	}
	papers := body["papers"].([]any)
	if len(papers) != len(wantPapers) || papers[0].(map[string]any)["duration_minutes"].(float64) <= 0 {
		t.Fatalf("详情试卷异常: %v", papers)
	}

	// 不存在的科目与非法参数
	if status, body = doJSON(t, ts.Client(), http.MethodGet, itBaseURL+"/api/v1/subjects/99999", "", auth); status != http.StatusNotFound || body["code"] != "SUBJECT_NOT_FOUND" {
		t.Fatalf("未知科目 status = %d, body %v", status, body)
	}
	if status, body = doJSON(t, ts.Client(), http.MethodGet, itBaseURL+"/api/v1/subjects/abc", "", auth); status != http.StatusBadRequest || body["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("非法科目参数 status = %d, body %v", status, body)
	}
}

func TestChapterQuestionsOrder(t *testing.T) {
	ts, db := newBankTestServer(t)
	client := ts.Client()
	auth := bankToken(t, client)

	var subject model.Subject
	if err := db.Where("code = ?", "soft-designer").First(&subject).Error; err != nil {
		t.Fatalf("查种子科目: %v", err)
	}
	var chapter model.Chapter
	if err := db.Where("subject_id = ? AND code = ?", subject.ID, "computer-system").First(&chapter).Error; err != nil {
		t.Fatalf("查种子章节: %v", err)
	}
	var wantIDs []uint
	if err := db.Model(&model.Question{}).Where("chapter_id = ?", chapter.ID).Order("id").
		Pluck("id", &wantIDs).Error; err != nil || len(wantIDs) == 0 {
		t.Fatalf("查章节题目: %v", err)
	}

	url := itBaseURL + "/api/v1/subjects/" + itoa(subject.ID) + "/chapters/" + itoa(chapter.ID) + "/questions"
	status, body := doJSON(t, client, http.MethodGet, url, "", auth)
	if status != http.StatusOK {
		t.Fatalf("章节题目 status = %d, body %v", status, body)
	}
	items, _ := body["questions"].([]any)
	if len(items) != len(wantIDs) {
		t.Fatalf("章节题目数 = %d, want %d", len(items), len(wantIDs))
	}
	assertNoAnswerLeak(t, items)
	for i, item := range items {
		q := item.(map[string]any)
		if q["id"].(float64) != float64(wantIDs[i]) {
			t.Fatalf("顺序模式第 %d 题应按 ID 稳定排序: got %v", i, q["id"])
		}
		if q["type"] != model.QuestionTypeSingle && q["type"] != model.QuestionTypeMultiple {
			t.Fatalf("题型非法: %v", q["type"])
		}
		if len(q["options"].([]any)) == 0 {
			t.Fatalf("题目缺选项: %v", q)
		}
	}

	// 章节跨科目引用与不存在的章节均 404；无题章节返回空数组而非 null
	other := model.Subject{Code: "other-subject", Name: "其它科目"}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("建第二科目: %v", err)
	}
	otherChapter := model.Chapter{SubjectID: other.ID, Code: "oc", Name: "其它章节", Sort: 1}
	if err := db.Create(&otherChapter).Error; err != nil {
		t.Fatalf("建第二章节: %v", err)
	}

	url = itBaseURL + "/api/v1/subjects/" + itoa(subject.ID) + "/chapters/" + itoa(otherChapter.ID) + "/questions"
	if status, body = doJSON(t, client, http.MethodGet, url, "", auth); status != http.StatusNotFound || body["code"] != "CHAPTER_NOT_FOUND" {
		t.Fatalf("跨科目章节 status = %d, body %v", status, body)
	}
	url = itBaseURL + "/api/v1/subjects/" + itoa(subject.ID) + "/chapters/99999/questions"
	if status, body = doJSON(t, client, http.MethodGet, url, "", auth); status != http.StatusNotFound {
		t.Fatalf("未知章节 status = %d, body %v", status, body)
	}

	url = itBaseURL + "/api/v1/subjects/" + itoa(other.ID) + "/chapters/" + itoa(otherChapter.ID) + "/questions"
	if status, body = doJSON(t, client, http.MethodGet, url, "", auth); status != http.StatusOK {
		t.Fatalf("空章节 status = %d, body %v", status, body)
	}
	if questions, ok := body["questions"].([]any); !ok || len(questions) != 0 {
		t.Fatalf("无题章节应返回空数组: %v", body["questions"])
	}
}

func TestRandomQuestions(t *testing.T) {
	ts, db := newBankTestServer(t)
	client := ts.Client()
	auth := bankToken(t, client)

	var subject model.Subject
	if err := db.Where("code = ?", "soft-designer").First(&subject).Error; err != nil {
		t.Fatalf("查种子科目: %v", err)
	}
	var wantSubjectIDs []uint
	if err := db.Model(&model.Question{}).Where("subject_id = ?", subject.ID).
		Pluck("id", &wantSubjectIDs).Error; err != nil || len(wantSubjectIDs) == 0 {
		t.Fatalf("查科目题目: %v", err)
	}

	// 全科目随机：元素集合与 DB 一致（不丢不重），顺序由打乱单测保证
	url := itBaseURL + "/api/v1/subjects/" + itoa(subject.ID) + "/questions/random"
	status, body := doJSON(t, client, http.MethodGet, url, "", auth)
	if status != http.StatusOK {
		t.Fatalf("随机取题 status = %d, body %v", status, body)
	}
	items := body["questions"].([]any)
	assertNoAnswerLeak(t, items)
	got := questionIDsFromResp(t, items)
	if len(got) != len(wantSubjectIDs) {
		t.Fatalf("随机题数 = %d, want %d", len(got), len(wantSubjectIDs))
	}
	for i, id := range got {
		if id != float64(wantSubjectIDs[i]) {
			t.Fatalf("随机题目集合与 DB 不符: got[%d] = %v", i, id)
		}
	}

	// 限定章节随机：集合等于该章节顺序取题集合
	var chapter model.Chapter
	if err := db.Where("subject_id = ? AND code = ?", subject.ID, "computer-system").First(&chapter).Error; err != nil {
		t.Fatalf("查种子章节: %v", err)
	}
	var wantChapterIDs []uint
	if err := db.Model(&model.Question{}).Where("chapter_id = ?", chapter.ID).Order("id").
		Pluck("id", &wantChapterIDs).Error; err != nil || len(wantChapterIDs) == 0 {
		t.Fatalf("查章节题目: %v", err)
	}

	url = itBaseURL + "/api/v1/subjects/" + itoa(subject.ID) + "/questions/random?chapter_id=" + itoa(chapter.ID)
	if status, body = doJSON(t, client, http.MethodGet, url, "", auth); status != http.StatusOK {
		t.Fatalf("限定章节随机 status = %d, body %v", status, body)
	}
	got = questionIDsFromResp(t, body["questions"].([]any))
	if len(got) != len(wantChapterIDs) {
		t.Fatalf("限定章节随机题数 = %d, want %d", len(got), len(wantChapterIDs))
	}
	for i, id := range got {
		if id != float64(wantChapterIDs[i]) {
			t.Fatalf("限定章节随机集合不符: got[%d] = %v, want %v", i, id, wantChapterIDs[i])
		}
	}

	// 参数错误：未知科目 404、未知章节 404、非法 chapter_id 400
	if status, body = doJSON(t, client, http.MethodGet, itBaseURL+"/api/v1/subjects/99999/questions/random", "", auth); status != http.StatusNotFound {
		t.Fatalf("未知科目随机 status = %d, body %v", status, body)
	}
	if status, body = doJSON(t, client, http.MethodGet, itBaseURL+"/api/v1/subjects/"+itoa(subject.ID)+"/questions/random?chapter_id=99999", "", auth); status != http.StatusNotFound {
		t.Fatalf("未知章节随机 status = %d, body %v", status, body)
	}
	if status, body = doJSON(t, client, http.MethodGet, itBaseURL+"/api/v1/subjects/"+itoa(subject.ID)+"/questions/random?chapter_id=abc", "", auth); status != http.StatusBadRequest {
		t.Fatalf("非法 chapter_id status = %d, body %v", status, body)
	}
}

func TestPaperListAndDetail(t *testing.T) {
	ts, db := newBankTestServer(t)
	client := ts.Client()
	auth := bankToken(t, client)

	var wantPapers []model.Paper
	if err := db.Order("id").Find(&wantPapers).Error; err != nil || len(wantPapers) == 0 {
		t.Fatalf("查试卷: %v", err)
	}

	status, body := doJSON(t, client, http.MethodGet, itBaseURL+"/api/v1/papers", "", auth)
	if status != http.StatusOK {
		t.Fatalf("试卷列表 status = %d, body %v", status, body)
	}
	papers := body["papers"].([]any)
	if len(papers) != len(wantPapers) {
		t.Fatalf("试卷数 = %d, want %d", len(papers), len(wantPapers))
	}

	paper := wantPapers[0]
	var wantNos []int
	if err := db.Model(&model.PaperQuestion{}).Where("paper_id = ?", paper.ID).Order("no").
		Pluck("no", &wantNos).Error; err != nil || len(wantNos) == 0 {
		t.Fatalf("查试卷题目: %v", err)
	}

	// 整卷详情：题号升序、数量一致；原始响应体不含答案与解析字段
	url := itBaseURL + "/api/v1/papers/" + itoa(paper.ID)
	status, body = doJSON(t, client, http.MethodGet, url, "", auth)
	if status != http.StatusOK {
		t.Fatalf("试卷详情 status = %d, body %v", status, body)
	}
	if body["paper"].(map[string]any)["duration_minutes"].(float64) != float64(paper.DurationMinutes) {
		t.Fatalf("试卷时长不符: %v", body["paper"])
	}
	items := body["questions"].([]any)
	assertNoAnswerLeak(t, items)
	if len(items) != len(wantNos) {
		t.Fatalf("整卷题数 = %d, want %d", len(items), len(wantNos))
	}
	for i, item := range items {
		q := item.(map[string]any)
		if q["no"].(float64) != float64(wantNos[i]) {
			t.Fatalf("整卷题号未按卷内顺序: got[%d] = %v, want %v", i, q["no"], wantNos[i])
		}
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("构造请求: %v", err)
	}
	req.Header.Set("Authorization", auth)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("请求试卷详情: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(raw), `"answer"`) || strings.Contains(string(raw), `"analysis"`) {
		t.Fatalf("整卷响应泄露答案/解析字段: %s", raw)
	}

	if status, body = doJSON(t, client, http.MethodGet, itBaseURL+"/api/v1/papers/99999", "", auth); status != http.StatusNotFound || body["code"] != "PAPER_NOT_FOUND" {
		t.Fatalf("未知试卷 status = %d, body %v", status, body)
	}
}

// itoa uint 转字符串的小助手
func itoa(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}
