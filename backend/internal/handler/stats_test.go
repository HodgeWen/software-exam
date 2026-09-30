package handler

import (
	"net/http"
	"testing"
)

// TestStatsSummary 统计接口：基于答题记录返回总量、正确率与按科目/章节的分布；按用户隔离
func TestStatsSummary(t *testing.T) {
	ts, db := newAnswerTestServer(t)
	client := ts.Client()
	f := seedAnswerFixtures(t, db)
	alice := userToken(t, client, "alice")
	bob := userToken(t, client, "bob")
	answersURL := itBaseURL + "/api/v1/answers"
	statsURL := itBaseURL + "/api/v1/stats"

	// 未登录 401
	if status, body := doJSON(t, client, http.MethodGet, statsURL, "", ""); status != http.StatusUnauthorized || body["code"] != "UNAUTHORIZED" {
		t.Fatalf("未登录查统计 status = %d, body %v", status, body)
	}

	// alice：A 科章节一 1 答 1 对，章节二 2 答 1 对，B 科（真题题，无章节）1 答 1 对
	doJSON(t, client, http.MethodPost, answersURL, submitBody(f.single.ID, `"A"`), alice)
	doJSON(t, client, http.MethodPost, answersURL, submitBody(f.multi.ID, `"A"`), alice)
	doJSON(t, client, http.MethodPost, answersURL, submitBody(f.multi.ID, `"A","C"`), alice)
	doJSON(t, client, http.MethodPost, answersURL, submitBody(f.other.ID, `"B"`), alice)

	status, body := doJSON(t, client, http.MethodGet, statsURL, "", alice)
	if status != http.StatusOK {
		t.Fatalf("查统计 status = %d, body %v", status, body)
	}
	if body["total"] != float64(4) || body["correct"] != float64(3) || body["accuracy"] != 0.75 {
		t.Fatalf("总量/正确率错误: %v", body)
	}
	subs, _ := body["subjects"].([]any)
	if len(subs) != 2 {
		t.Fatalf("科目数 = %d, want 2: %v", len(subs), body)
	}
	a, _ := subs[0].(map[string]any)
	if a["subject_id"] != float64(f.subjA.ID) || a["subject_name"] != "科目A" ||
		a["total"] != float64(3) || a["correct"] != float64(2) || a["accuracy"] != 0.6667 {
		t.Fatalf("科目 A 统计错误: %v", a)
	}
	chs, _ := a["chapters"].([]any)
	if len(chs) != 2 {
		t.Fatalf("科目 A 章节数 = %d, want 2: %v", len(chs), a)
	}
	ch1, _ := chs[0].(map[string]any)
	if ch1["chapter_id"] != float64(f.chA1.ID) || ch1["chapter_name"] != "章节一" ||
		ch1["total"] != float64(1) || ch1["correct"] != float64(1) || ch1["accuracy"] != float64(1) {
		t.Fatalf("章节一统计错误: %v", ch1)
	}
	ch2, _ := chs[1].(map[string]any)
	if ch2["chapter_id"] != float64(f.chA2.ID) || ch2["total"] != float64(2) ||
		ch2["correct"] != float64(1) || ch2["accuracy"] != 0.5 {
		t.Fatalf("章节二统计错误: %v", ch2)
	}
	b, _ := subs[1].(map[string]any)
	if b["subject_id"] != float64(f.subjB.ID) || b["total"] != float64(1) || b["correct"] != float64(1) {
		t.Fatalf("科目 B 统计错误: %v", b)
	}
	if bchs, _ := b["chapters"].([]any); len(bchs) != 0 {
		t.Fatalf("真题题只计科目不进章节分布: %v", b)
	}

	// 用户隔离：bob 无作答，全 0
	status, body = doJSON(t, client, http.MethodGet, statsURL, "", bob)
	if status != http.StatusOK || body["total"] != float64(0) || body["accuracy"] != float64(0) {
		t.Fatalf("他人统计不应可见 status = %d, body %v", status, body)
	}
	if subs, _ := body["subjects"].([]any); len(subs) != 0 {
		t.Fatalf("无作答应无科目分布: %v", body)
	}
}
