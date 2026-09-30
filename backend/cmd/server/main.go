package main

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"software-exam/backend/internal/handler"
	"software-exam/backend/internal/model"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	gin.SetMode(gin.ReleaseMode)
	if err := run(); err != nil {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func run() error {
	dbPath := envOr("EXAM_DB_PATH", "exam.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{TranslateError: true})
	if err != nil {
		return fmt.Errorf("打开 SQLite %s: %w", dbPath, err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return fmt.Errorf("迁移数据库: %w", err)
	}

	secret, err := jwtSecret()
	if err != nil {
		return err
	}

	addr := ":" + envOr("EXAM_PORT", "8080")
	slog.Info("listening", "addr", addr, "db", dbPath)
	return handler.NewRouter(db, secret).Run(addr)
}

// jwtSecret 优先读环境变量；未配置则生成进程内随机密钥（安全默认，代价是重启后旧 token 失效）
func jwtSecret() ([]byte, error) {
	if s := os.Getenv("EXAM_JWT_SECRET"); s != "" {
		return []byte(s), nil
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("生成 JWT 密钥: %w", err)
	}
	slog.Warn("EXAM_JWT_SECRET 未配置，使用随机密钥，重启后已签发 token 将失效")
	return buf, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
