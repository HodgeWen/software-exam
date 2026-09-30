package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/service"
)

var mwSecret = []byte("middleware-test-secret")

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	u := &model.User{ID: 42, Username: "alice"}
	valid, err := service.IssueToken(mwSecret, u, time.Hour)
	if err != nil {
		t.Fatalf("签发 token: %v", err)
	}
	expired, err := service.IssueToken(mwSecret, u, -time.Minute)
	if err != nil {
		t.Fatalf("签发过期 token: %v", err)
	}

	r := gin.New()
	r.Use(JWTAuth(mwSecret))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":       c.GetUint(CtxUserID),
			"username": c.GetString(CtxUsername),
		})
	})

	tests := []struct {
		name     string
		auth     string
		wantStat int
	}{
		{"无认证头", "", http.StatusUnauthorized},
		{"非 Bearer 方案", "Token " + valid, http.StatusUnauthorized},
		{"Bearer 后为空", "Bearer ", http.StatusUnauthorized},
		{"无效 token", "Bearer garbage.token.sig", http.StatusUnauthorized},
		{"过期 token", "Bearer " + expired, http.StatusUnauthorized},
		{"有效 token", "Bearer " + valid, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStat {
				t.Fatalf("status = %d, want %d, body %s", w.Code, tt.wantStat, w.Body)
			}
		})
	}

	// 有效 token 时用户信息应写入上下文
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+valid)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应: %v", err)
	}
	if body["id"] != float64(42) || body["username"] != "alice" {
		t.Fatalf("上下文用户不符: %v", body)
	}
}
