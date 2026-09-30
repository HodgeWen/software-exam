package handler

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// favoriteBody 收藏请求体
func favoriteBody(questionID uint) string {
	return fmt.Sprintf(`{"question_id":%d}`, questionID)
}

// postNoContent 发带 JSON 体的请求并断言 204（无响应体）
func postNoContent(t *testing.T, client *http.Client, url, body, auth string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("构造请求: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("POST %s status = %d, want 204", url, resp.StatusCode)
	}
}

// favoriteItems 把收藏列表响应按题目 ID 索引，便于跨顺序断言
func favoriteItems(t *testing.T, body map[string]any) map[uint]map[string]any {
	t.Helper()
	list, _ := body["list"].([]any)
	items := make(map[uint]map[string]any, len(list))
	for _, it := range list {
		item, ok := it.(map[string]any)
		if !ok {
			t.Fatalf("收藏行不是对象: %v", it)
		}
		q, ok := item["question"].(map[string]any)
		if !ok {
			t.Fatalf("收藏行缺少题目对象: %v", item)
		}
		items[uint(q["id"].(float64))] = item
	}
	return items
}

// TestFavoriteFlow 收藏：鉴权与参数校验、幂等收藏/取消、分页列表（含 total）、按用户隔离
func TestFavoriteFlow(t *testing.T) {
	ts, db := newAnswerTestServer(t)
	client := ts.Client()
	f := seedAnswerFixtures(t, db)
	alice := userToken(t, client, "alice")
	bob := userToken(t, client, "bob")
	favoritesURL := itBaseURL + "/api/v1/favorites"

	// 未登录 401
	if status, body := doJSON(t, client, http.MethodPost, favoritesURL,
		favoriteBody(f.single.ID), ""); status != http.StatusUnauthorized || body["code"] != "UNAUTHORIZED" {
		t.Fatalf("未登录收藏 status = %d, body %v", status, body)
	}
	// 缺题目 ID 400
	if status, body := doJSON(t, client, http.MethodPost, favoritesURL, `{}`, alice); status != http.StatusBadRequest || body["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("缺题目 ID status = %d, body %v", status, body)
	}

	// alice 收藏两题；bob 收藏另一题
	postNoContent(t, client, favoritesURL, favoriteBody(f.single.ID), alice)
	postNoContent(t, client, favoritesURL, favoriteBody(f.multi.ID), alice)
	postNoContent(t, client, favoritesURL, favoriteBody(f.other.ID), bob)
	// 重复收藏：仍 204，不产生重复记录
	postNoContent(t, client, favoritesURL, favoriteBody(f.single.ID), alice)

	status, body := doJSON(t, client, http.MethodGet, favoritesURL, "", alice)
	if status != http.StatusOK || body["total"] != float64(2) {
		t.Fatalf("收藏列表 status = %d, body %v", status, body)
	}
	items := favoriteItems(t, body)
	if len(items) != 2 {
		t.Fatalf("收藏行数 = %d, want 2（重复收藏不得新增行）", len(items))
	}
	item := items[f.single.ID]
	if item == nil {
		t.Fatalf("收藏列表应含题目 %d: %v", f.single.ID, items)
	}
	if item["created_at"] == nil {
		t.Fatalf("收藏行应含收藏时间: %v", item)
	}
	if _, has := item["question"].(map[string]any)["answer"]; has {
		t.Fatalf("收藏列表泄露正确答案: %v", item)
	}

	// 分页：响应含 total，每页条数生效
	status, body = doJSON(t, client, http.MethodGet, favoritesURL+"?page_size=1&page=2", "", alice)
	if status != http.StatusOK || body["total"] != float64(2) || len(body["list"].([]any)) != 1 {
		t.Fatalf("收藏分页 status = %d, body %v", status, body)
	}

	// 用户隔离：bob 只见自己的收藏
	status, body = doJSON(t, client, http.MethodGet, favoritesURL, "", bob)
	if status != http.StatusOK || body["total"] != float64(1) {
		t.Fatalf("他人收藏不应可见 status = %d, body %v", status, body)
	}
	if _, has := favoriteItems(t, body)[f.other.ID]; !has {
		t.Fatalf("bob 收藏列表应含题目 %d", f.other.ID)
	}

	// 取消收藏：204，幂等，且不影响他人
	removeURL := fmt.Sprintf("%s/%d", favoritesURL, f.single.ID)
	doNoContent(t, client, http.MethodDelete, removeURL, alice)
	doNoContent(t, client, http.MethodDelete, removeURL, alice)
	status, body = doJSON(t, client, http.MethodGet, favoritesURL, "", alice)
	if body["total"] != float64(1) || len(favoriteItems(t, body)) != 1 {
		t.Fatalf("取消后应剩 1 条收藏: %v", body)
	}
	// bob 取消 alice 仍持有的收藏：返回成功但不影响 alice
	doNoContent(t, client, http.MethodDelete, fmt.Sprintf("%s/%d", favoritesURL, f.multi.ID), bob)
	status, body = doJSON(t, client, http.MethodGet, favoritesURL, "", alice)
	if body["total"] != float64(1) {
		t.Fatalf("他人取消不应影响本人收藏: %v", body)
	}

	// 路径参数不合法 400
	if status, body = doJSON(t, client, http.MethodDelete, favoritesURL+"/abc", "", alice); status != http.StatusBadRequest || body["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("非法题目 ID status = %d, body %v", status, body)
	}
}
