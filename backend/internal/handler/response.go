package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CodeInvalidArgument 请求参数校验失败的通用业务错误码
const CodeInvalidArgument = "INVALID_ARGUMENT"

// Error 统一错误响应：{"code": "<业务错误码>", "message": "<人读信息>"}
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"code": code, "message": message})
}

// bindJSON 绑定并校验 JSON 请求体，失败时已写好 400 响应，返回 false 表示终止
func bindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		Error(c, http.StatusBadRequest, CodeInvalidArgument, "请求参数错误: "+err.Error())
		return false
	}
	return true
}

func internalError(c *gin.Context, err error) {
	slog.Error("internal error", "error", err, "path", c.Request.URL.Path)
	Error(c, http.StatusInternalServerError, "INTERNAL", "服务器内部错误")
}
