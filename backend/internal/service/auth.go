package service

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/repository"
)

// 密码哈希用 PBKDF2-HMAC-SHA256（标准库实现），存储格式：
// pbkdf2-sha256$<迭代次数>$<盐 hex>$<派生键 hex>
const (
	pbkdf2Iterations = 600_000
	pbkdf2KeyLength  = 32
	tokenTTL         = 24 * time.Hour
)

var (
	ErrUsernameExists     = errors.New("用户名已存在")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
)

// Claims JWT 载荷；uid 供后续阶段做业务数据按用户隔离
type Claims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type AuthService struct {
	repo   *repository.UserRepository
	secret []byte
}

func NewAuthService(repo *repository.UserRepository, secret []byte) *AuthService {
	return &AuthService{repo: repo, secret: secret}
}

// Register 创建账号；用户名冲突（含唯一约束兜底）返回 ErrUsernameExists
func (s *AuthService) Register(username, password string) (*model.User, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &model.User{Username: username, Password: hash}
	if err := s.repo.Create(u); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrUsernameExists
		}
		return nil, fmt.Errorf("创建用户: %w", err)
	}
	return u, nil
}

// Login 校验凭据并签发 JWT；用户不存在与密码错误统一返回 ErrInvalidCredentials，避免探测用户名
func (s *AuthService) Login(username, password string) (string, *model.User, error) {
	u, err := s.repo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("查询用户: %w", err)
	}
	if !verifyPassword(u.Password, password) {
		return "", nil, ErrInvalidCredentials
	}
	token, err := IssueToken(s.secret, u, tokenTTL)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

// IssueToken 以 HS256 签发 JWT；ttl 为负数签出的即过期令牌（测试用）
func IssueToken(secret []byte, u *model.User, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   u.ID,
		Username: u.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// ParseToken 校验签名与有效期并还原载荷
func ParseToken(secret []byte, token string) (*Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	return &claims, nil
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("生成盐值: %w", err)
	}
	dk, err := pbkdf2.Key(sha256.New, password, salt, pbkdf2Iterations, pbkdf2KeyLength)
	if err != nil {
		return "", fmt.Errorf("派生密码哈希: %w", err)
	}
	return "pbkdf2-sha256$" + strconv.Itoa(pbkdf2Iterations) + "$" +
		hex.EncodeToString(salt) + "$" + hex.EncodeToString(dk), nil
}

func verifyPassword(stored, password string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter <= 0 {
		return false
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(parts[3])
	if err != nil {
		return false
	}
	dk, err := pbkdf2.Key(sha256.New, password, salt, iter, len(want))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(dk, want) == 1
}
