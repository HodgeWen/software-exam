package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/repository"
)

var testSecret = []byte("unit-test-secret")

func newAuthService(t *testing.T) *AuthService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("打开内存 SQLite: %v", err)
	}
	// 内存库每个连接独立，限制单连接避免「no such table」
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("取底层 sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("迁移: %v", err)
	}
	return NewAuthService(repository.NewUserRepository(db), testSecret)
}

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := hashPassword("s3cret!")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if strings.Contains(hash, "s3cret!") || hash == "" {
		t.Fatalf("哈希包含明文或为空: %q", hash)
	}
	if !strings.HasPrefix(hash, "pbkdf2-sha256$") {
		t.Fatalf("哈希格式不符: %q", hash)
	}
	if !verifyPassword(hash, "s3cret!") {
		t.Error("正确密码校验失败")
	}
	if verifyPassword(hash, "wrong") {
		t.Error("错误密码校验通过")
	}
	// 随机盐：同一密码两次哈希不同，且都能通过校验
	again, _ := hashPassword("s3cret!")
	if again == hash {
		t.Error("两次哈希相同，盐可能未随机")
	}
	if !verifyPassword(again, "s3cret!") {
		t.Error("第二次哈希校验失败")
	}
	if verifyPassword("not-a-hash", "s3cret!") {
		t.Error("非法存储格式校验通过")
	}
}

func TestIssueAndParseToken(t *testing.T) {
	u := &model.User{ID: 7, Username: "alice"}
	token, err := IssueToken(testSecret, u, time.Hour)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	claims, err := ParseToken(testSecret, token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != 7 || claims.Username != "alice" {
		t.Fatalf("载荷不符: %+v", claims)
	}
	if _, err := ParseToken([]byte("wrong-secret"), token); err == nil {
		t.Error("错误密钥解析通过")
	}
	if _, err := ParseToken(testSecret, "garbage.token"); err == nil {
		t.Error("垃圾串解析通过")
	}
	// 过期令牌应报 ErrTokenExpired
	expired, err := IssueToken(testSecret, u, -time.Minute)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if _, err := ParseToken(testSecret, expired); !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("过期令牌未报 ErrTokenExpired: %v", err)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	s := newAuthService(t)

	u, err := s.Register("alice", "pass123")
	if err != nil {
		t.Fatalf("注册: %v", err)
	}
	if u.ID == 0 || u.Password == "pass123" || !strings.HasPrefix(u.Password, "pbkdf2-sha256$") {
		t.Fatalf("返回用户异常: id=%d password=%q", u.ID, u.Password)
	}

	if _, err := s.Register("alice", "other456"); !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("重名注册未返回 ErrUsernameExists: %v", err)
	}

	if _, _, err := s.Login("alice", "nope"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("错误密码未返回 ErrInvalidCredentials: %v", err)
	}
	if _, _, err := s.Login("nobody", "pass123"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("未知用户未返回 ErrInvalidCredentials: %v", err)
	}

	token, got, err := s.Login("alice", "pass123")
	if err != nil {
		t.Fatalf("登录: %v", err)
	}
	claims, err := ParseToken(testSecret, token)
	if err != nil {
		t.Fatalf("解析登录 token: %v", err)
	}
	if claims.UserID != got.ID || claims.Username != "alice" {
		t.Fatalf("登录 token 载荷不符: %+v", claims)
	}
}
