package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"software-exam/backend/internal/model"
)

// newStaticTestServer 起 minimal 服务并托管临时静态目录，验证 SPA 托管行为
func newStaticTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	files := map[string]string{
		"index.html":          "<!doctype html><title>exam-spa</title>",
		"assets/app.js":       "console.log(1)",
		"assets/nested/c.css": "body{}",
	}
	for name, content := range files {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("建目录: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("写文件: %v", err)
		}
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("打开内存 SQLite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("迁移: %v", err)
	}
	ts := httptest.NewServer(NewRouter(db, itSecret, dir))
	t.Cleanup(ts.Close)
	return ts
}

func TestStaticHosting(t *testing.T) {
	ts := newStaticTestServer(t)
	client := ts.Client()

	// 命中静态文件：原样返回
	for path, want := range map[string]string{
		"/":                    "<!doctype html><title>exam-spa</title>",
		"/index.html":          "<!doctype html><title>exam-spa</title>",
		"/assets/app.js":       "console.log(1)",
		"/assets/nested/c.css": "body{}",
	} {
		resp, err := client.Get(ts.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || string(raw) != want {
			t.Fatalf("GET %s = %d %q, want 200 %q", path, resp.StatusCode, raw, want)
		}
	}

	// 前端路由路径：回落 index.html
	resp, err := client.Get(ts.URL + "/subjects/3/chapters/5/browse")
	if err != nil {
		t.Fatalf("GET SPA 路由: %v", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || string(raw) != "<!doctype html><title>exam-spa</title>" {
		t.Fatalf("SPA 回落异常: %d %q", resp.StatusCode, raw)
	}

	// /api 未知路径：JSON 404 而非 index.html
	resp, err = client.Get(ts.URL + "/api/v1/nope")
	if err != nil {
		t.Fatalf("GET 未知 API: %v", err)
	}
	raw, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound || string(raw) == "" {
		t.Fatalf("未知 API 应 404: %d %q", resp.StatusCode, raw)
	}

	// 目录访问：回落 index.html，不列目录
	resp, err = client.Get(ts.URL + "/assets")
	if err != nil {
		t.Fatalf("GET /assets: %v", err)
	}
	raw, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || string(raw) != "<!doctype html><title>exam-spa</title>" {
		t.Fatalf("GET /assets 应回落 index.html: %d %q", resp.StatusCode, raw)
	}
}

// TestFileExistsResistsTraversal 目录穿越只会在静态目录内探测，永远不会越出
func TestFileExistsResistsTraversal(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("s"), 0o644); err != nil {
		t.Fatalf("写外部文件: %v", err)
	}
	// 用外部目录名做静态目录的兄弟目录，尝试 ../<外部目录>/secret.txt 逃逸
	rel := "../" + filepath.Base(outside) + "/secret.txt"
	if fileExists(dir, "/"+rel) {
		t.Fatalf("目录穿越命中静态目录外文件: %s", rel)
	}
}
