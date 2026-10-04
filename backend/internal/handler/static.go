package handler

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// serveStatic 托管前端 SPA 静态资源（生产部署形态：单进程同时提供 API 与前端页面）。
// 静态命中的路径直接返回文件；其余非 /api 路径一律回落 index.html，保证前端路由刷新可直出。
// 路径先规范化再拼接到根目录内，防止目录穿越。
func serveStatic(r *gin.Engine, staticDir string) {
	fileServer := http.FileServer(http.Dir(staticDir))
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			Error(c, http.StatusNotFound, "NOT_FOUND", "接口不存在")
			return
		}
		if fileExists(staticDir, c.Request.URL.Path) {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.File(filepath.Join(staticDir, "index.html"))
	})
}

// fileExists 判断 URL 路径是否映射到静态目录内的真实文件（目录不算，回落 index.html）
func fileExists(staticDir, urlPath string) bool {
	rel := strings.TrimPrefix(path.Clean("/"+urlPath), "/")
	if rel == "" {
		return false
	}
	st, err := os.Stat(filepath.Join(staticDir, rel))
	return err == nil && !st.IsDir()
}
