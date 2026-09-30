package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health GET /api/v1/health 存活探针（公开）
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
