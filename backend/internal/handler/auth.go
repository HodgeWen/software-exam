package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"software-exam/backend/internal/middleware"
	"software-exam/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// credentials 注册与登录共用同一组字段，无需两份 DTO
type credentials struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type userResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Register POST /api/v1/auth/register（公开）
func (h *AuthHandler) Register(c *gin.Context) {
	var req credentials
	if !bindJSON(c, &req) {
		return
	}
	u, err := h.auth.Register(req.Username, req.Password)
	switch {
	case errors.Is(err, service.ErrUsernameExists):
		Error(c, http.StatusConflict, "USERNAME_EXISTS", "用户名已存在")
	case err != nil:
		internalError(c, err)
	default:
		c.JSON(http.StatusCreated, userResponse{ID: u.ID, Username: u.Username, CreatedAt: u.CreatedAt})
	}
}

// Login POST /api/v1/auth/login（公开）
func (h *AuthHandler) Login(c *gin.Context) {
	var req credentials
	if !bindJSON(c, &req) {
		return
	}
	token, u, err := h.auth.Login(req.Username, req.Password)
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误")
	case err != nil:
		internalError(c, err)
	default:
		c.JSON(http.StatusOK, gin.H{"token": token, "user": userResponse{ID: u.ID, Username: u.Username, CreatedAt: u.CreatedAt}})
	}
}

// Me GET /api/v1/me（鉴权）：返回当前登录用户，供前端会话引导与 token 校验
func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":       c.GetUint(middleware.CtxUserID),
		"username": c.GetString(middleware.CtxUsername),
	})
}
