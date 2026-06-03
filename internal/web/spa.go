package web

import (
	"io/fs"
	"net/http"
	"os"
	"strings"
)

// 前端构建产物目录（开发时可能不存在）
// 构建时通过 `go build -tags withweb` 启用嵌入
var assets fs.FS

func init() {
	// 开发模式：从本地文件系统读取前端资源
	// 生产模式：通过 `go build -tags withweb` 编译时嵌入
	assets = os.DirFS("internal/web/dist")
}

// SPAHandler 服务前端静态资源并提供 SPA fallback
type SPAHandler struct {
	fileServer http.Handler
}

// NewSPAHandler 创建 SPA handler
func NewSPAHandler() *SPAHandler {
	return &SPAHandler{
		fileServer: http.FileServer(http.FS(assets)),
	}
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	// API 路径不处理
	if strings.HasPrefix(path, "api/") || strings.HasPrefix(path, "d/") {
		http.NotFound(w, r)
		return
	}

	// 尝试查找静态资源
	if path != "" {
		statPath := path
		f, err := assets.Open(statPath)
		if err == nil {
			f.Close()
			h.fileServer.ServeHTTP(w, r)
			return
		}
	}

	// SPA fallback: 返回 index.html
	r.URL.Path = "/"
	h.fileServer.ServeHTTP(w, r)
}