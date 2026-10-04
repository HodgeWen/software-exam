package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/service"
)

var (
	itSecret = []byte("integration-test-secret")
	// 内存假网络下任意地址都会路由到测试服务器，统一用一个固定 base
	itBaseURL = "http://example.com"
)

// newTestServer 用内存 SQLite + 真实路由起 httptest 服务，走完整 HTTP 栈
func newTestServer(t *testing.T) (*httptest.Server, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("打开内存 SQLite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("取底层 sql.DB: %v", err)
	}
	// 内存库每个连接独立，限制单连接避免「no such table」
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("迁移: %v", err)
	}
	ts := httptest.NewTestServer(t, NewRouter(db, itSecret, ""))
	return ts, db
}

func doJSON(t *testing.T, client *http.Client, method, url, body, auth string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("构造请求: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("解析响应: %v", err)
	}
	return resp.StatusCode, got
}

func TestHealth(t *testing.T) {
	ts, _ := newTestServer(t)
	status, body := doJSON(t, ts.Client(), http.MethodGet, itBaseURL+"/api/v1/health", "", "")
	if status != http.StatusOK || body["status"] != "ok" {
		t.Fatalf("health status = %d, body %v", status, body)
	}
}

func TestRegisterLoginAuthFlow(t *testing.T) {
	ts, db := newTestServer(t)
	client := ts.Client()
	registerURL := itBaseURL + "/api/v1/auth/register"
	loginURL := itBaseURL + "/api/v1/auth/login"
	meURL := itBaseURL + "/api/v1/me"

	// 注册成功：返回用户信息，不含密码字段
	status, body := doJSON(t, client, http.MethodPost, registerURL, `{"username":"alice","password":"pass123"}`, "")
	if status != http.StatusCreated {
		t.Fatalf("注册 status = %d, body %v", status, body)
	}
	if body["username"] != "alice" || body["id"].(float64) <= 0 {
		t.Fatalf("注册返回用户异常: %v", body)
	}
	if _, has := body["password"]; has {
		t.Fatalf("注册响应泄露密码字段: %v", body)
	}

	// 密码不以明文落库
	var stored []string
	if err := db.Model(&model.User{}).Where("username = ?", "alice").
		Pluck("password", &stored).Error; err != nil || len(stored) != 1 {
		t.Fatalf("查询落库密码失败: %v", err)
	}
	if stored[0] == "pass123" || !strings.HasPrefix(stored[0], "pbkdf2-sha256$") {
		t.Fatalf("落库密码疑似明文: %q", stored[0])
	}

	// 重名注册：4xx + 统一错误体
	status, body = doJSON(t, client, http.MethodPost, registerURL, `{"username":"alice","password":"other456"}`, "")
	if status < 400 || status >= 500 {
		t.Fatalf("重名注册 status = %d, body %v", status, body)
	}
	if body["code"] != "USERNAME_EXISTS" || body["message"] == "" {
		t.Fatalf("重名注册错误体不符: %v", body)
	}

	// 参数缺失：400 + 统一错误体
	status, body = doJSON(t, client, http.MethodPost, registerURL, `{"username":"bob"}`, "")
	if status != http.StatusBadRequest || body["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("缺参注册 status = %d, body %v", status, body)
	}

	// 错误凭据：401
	status, body = doJSON(t, client, http.MethodPost, loginURL, `{"username":"alice","password":"wrongpass"}`, "")
	if status != http.StatusUnauthorized || body["code"] != "INVALID_CREDENTIALS" {
		t.Fatalf("错凭据登录 status = %d, body %v", status, body)
	}

	// 登录成功：返回可解析出该用户的 JWT
	status, body = doJSON(t, client, http.MethodPost, loginURL, `{"username":"alice","password":"pass123"}`, "")
	if status != http.StatusOK {
		t.Fatalf("登录 status = %d, body %v", status, body)
	}
	token, _ := body["token"].(string)
	if token == "" {
		t.Fatalf("登录未返回 token: %v", body)
	}
	claims, err := service.ParseToken(itSecret, token)
	if err != nil || claims.Username != "alice" {
		t.Fatalf("登录 token 校验失败: %v %+v", err, claims)
	}

	// 受保护接口：无 token / 坏 token 401，有效 token 200
	if status, body = doJSON(t, client, http.MethodGet, meURL, "", ""); status != http.StatusUnauthorized || body["code"] != "UNAUTHORIZED" {
		t.Fatalf("无 token 访问 /me status = %d, body %v", status, body)
	}
	if status, body = doJSON(t, client, http.MethodGet, meURL, "", "Bearer not.a.jwt"); status != http.StatusUnauthorized {
		t.Fatalf("坏 token 访问 /me status = %d, body %v", status, body)
	}
	if status, body = doJSON(t, client, http.MethodGet, meURL, "", "Bearer "+token); status != http.StatusOK ||
		body["username"] != "alice" || body["id"].(float64) <= 0 {
		t.Fatalf("有效 token 访问 /me status = %d, body %v", status, body)
	}
}
