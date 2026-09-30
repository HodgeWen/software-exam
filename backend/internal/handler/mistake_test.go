package handler

import (
	"fmt"
	"net/http"
	"testing"
)

// mistakeItems 把错题列表响应按题目 ID 索引，便于跨顺序断言
func mistakeItems(t *testing.T, body map[string]any) map[uint]map[string]any {
	t.Helper()
	list, _ := body["list"].([]any)
	items := make(map[uint]map[string]any, len(list))
	for _, it := range list {
		item, ok := it.(map[string]any)
		if !ok {
			t.Fatalf("错题行不是对象: %v", it)
		}
		q, ok := item["question"].(map[string]any)
		if !ok {
			t.Fatalf("错题行缺少题目对象: %v", item)
		}
		items[uint(q["id"].(float64))] = item
	}
	return items
}

// doNoContent 请求无响应体接口（204），只断言状态码
func doNoContent(t *testing.T, client *http.Client, method, url, auth string) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("构造请求: %v", err)
	}
	req.Header.Set("Authorization", auth)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("%s %s status = %d, want 204", method, url, resp.StatusCode)
	}
}

// TestMistakeBookFlow 错题本：自动入库、科目过滤、分页、重刷题源、答对不移除、手动移除、用户隔离
func TestMistakeBookFlow(t *testing.T) {
	ts, db := newAnswerTestServer(t)
	client := ts.Client()
	f := seedAnswerFixtures(t, db)
	alice := userToken(t, client, "alice")
	bob := userToken(t, client, "bob")
	answersURL := itBaseURL + "/api/v1/answers"
	mistakesURL := itBaseURL + "/api/v1/mistakes"

	// 未登录 401
	if status, body := doJSON(t, client, http.MethodGet, mistakesURL, "", ""); status != http.StatusUnauthorized || body["code"] != "UNAUTHORIZED" {
		t.Fatalf("未登录查错题 status = %d, body %v", status, body)
	}
	// 空错题本
	status, body := doJSON(t, client, http.MethodGet, mistakesURL, "", alice)
	if status != http.StatusOK || body["total"] != float64(0) {
		t.Fatalf("初始错题本 status = %d, body %v", status, body)
	}

	// 单选答对不入错题；A、B 两科各答错一题
	doJSON(t, client, http.MethodPost, answersURL, submitBody(f.single.ID, `"A"`), alice)
	doJSON(t, client, http.MethodPost, answersURL, submitBody(f.multi.ID, `"A"`), alice)
	doJSON(t, client, http.MethodPost, answersURL, submitBody(f.other.ID, `"A"`), alice)

	status, body = doJSON(t, client, http.MethodGet, mistakesURL, "", alice)
	if status != http.StatusOK || body["total"] != float64(2) {
		t.Fatalf("两科各答错一题 status = %d, body %v", status, body)
	}
	items := mistakeItems(t, body)
	if len(items) != 2 {
		t.Fatalf("错题行数 = %d, want 2", len(items))
	}
	for qid, item := range items {
		if item["wrong_count"] != float64(1) {
			t.Fatalf("题目 %d 首次答错 wrong_count 应为 1: %v", qid, item)
		}
		if _, has := item["question"].(map[string]any)["answer"]; has {
			t.Fatalf("错题列表泄露正确答案: %v", item)
		}
	}

	// 按科目过滤
	status, body = doJSON(t, client, http.MethodGet,
		fmt.Sprintf("%s?subject_id=%d", mistakesURL, f.subjA.ID), "", alice)
	if status != http.StatusOK || body["total"] != float64(1) || len(body["list"].([]any)) != 1 {
		t.Fatalf("科目过滤 status = %d, body %v", status, body)
	}
	items = mistakeItems(t, body)
	if len(items) != 1 {
		t.Fatalf("科目过滤应只剩 A 科错题: %v", items)
	}
	if _, has := items[f.multi.ID]; !has {
		t.Fatalf("科目过滤应保留题目 %d: %v", f.multi.ID, items)
	}

	// 分页：响应含 total，每页条数生效
	status, body = doJSON(t, client, http.MethodGet, mistakesURL+"?page_size=1&page=1", "", alice)
	if status != http.StatusOK || body["total"] != float64(2) || len(body["list"].([]any)) != 1 {
		t.Fatalf("分页第 1 页 status = %d, body %v", status, body)
	}
	status, body = doJSON(t, client, http.MethodGet, mistakesURL+"?page_size=1&page=2", "", alice)
	if status != http.StatusOK || body["total"] != float64(2) || len(body["list"].([]any)) != 1 {
		t.Fatalf("分页第 2 页 status = %d, body %v", status, body)
	}

	// 重复答错不重复行：仅累加错误次数
	doJSON(t, client, http.MethodPost, answersURL, submitBody(f.multi.ID, `"B"`), alice)
	status, body = doJSON(t, client, http.MethodGet, mistakesURL, "", alice)
	if body["total"] != float64(2) {
		t.Fatalf("重复答错不应新增行: %v", body)
	}
	if item := mistakeItems(t, body)[f.multi.ID]; item["wrong_count"] != float64(2) {
		t.Fatalf("重复答错应累加错误次数: %v", item)
	}

	// 重刷答对：仅更新记录，不自动移除
	doJSON(t, client, http.MethodPost, answersURL, submitBody(f.multi.ID, `"A","C"`), alice)
	status, body = doJSON(t, client, http.MethodGet, mistakesURL, "", alice)
	if body["total"] != float64(2) {
		t.Fatalf("重刷答对不应移除错题: %v", body)
	}
	if item := mistakeItems(t, body)[f.multi.ID]; item["wrong_count"] != float64(2) {
		t.Fatalf("重刷答对不应累加错误次数: %v", item)
	}

	// 重刷题源：以错题为题源返回题目序列，不含正确答案与解析
	status, body = doJSON(t, client, http.MethodGet, mistakesURL+"/questions", "", alice)
	if status != http.StatusOK {
		t.Fatalf("错题重刷取题 status = %d, body %v", status, body)
	}
	questions, _ := body["questions"].([]any)
	if len(questions) != 2 {
		t.Fatalf("重刷题源应含 2 题: %v", body)
	}
	assertNoAnswerLeak(t, questions)

	// 用户隔离：bob 看不到 alice 的错题
	status, body = doJSON(t, client, http.MethodGet, mistakesURL, "", bob)
	if status != http.StatusOK || body["total"] != float64(0) {
		t.Fatalf("他人错题不应可见 status = %d, body %v", status, body)
	}

	// 手动移除：204，幂等，且不影响他人
	removeURL := fmt.Sprintf("%s/%d", mistakesURL, f.multi.ID)
	doNoContent(t, client, http.MethodDelete, removeURL, alice)
	status, body = doJSON(t, client, http.MethodGet, mistakesURL, "", alice)
	if body["total"] != float64(1) || len(mistakeItems(t, body)) != 1 {
		t.Fatalf("移除后应剩 1 条错题: %v", body)
	}
	if _, has := mistakeItems(t, body)[f.other.ID]; !has {
		t.Fatalf("移除的应是题目 %d: %v", f.multi.ID, body)
	}
	doNoContent(t, client, http.MethodDelete, removeURL, alice)
	// bob 移除 alice 仍持有的错题：返回成功但不影响 alice
	doNoContent(t, client, http.MethodDelete, fmt.Sprintf("%s/%d", mistakesURL, f.other.ID), bob)
	status, body = doJSON(t, client, http.MethodGet, mistakesURL, "", alice)
	if body["total"] != float64(1) {
		t.Fatalf("他人移除不应影响本人错题: %v", body)
	}

	// 路径参数不合法 400
	if status, body = doJSON(t, client, http.MethodDelete, mistakesURL+"/abc", "", alice); status != http.StatusBadRequest || body["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("非法题目 ID status = %d, body %v", status, body)
	}
}
