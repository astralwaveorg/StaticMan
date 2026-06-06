package web

import (
	"io/fs"
	"net/http"
	"os"
	"strings"
)

// SiteConfig 用于动态注入 index.html 的站点配置
type SiteConfig struct {
	Title       string
	Description string
	Logo        string
}

// SiteConfigFunc 获取当前站点配置的回调
type SiteConfigFunc func() SiteConfig

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
	getSite    SiteConfigFunc
}

// NewSPAHandler 创建 SPA handler
func NewSPAHandler(getSite ...SiteConfigFunc) *SPAHandler {
	h := &SPAHandler{
		fileServer: http.FileServer(http.FS(assets)),
	}
	if len(getSite) > 0 && getSite[0] != nil {
		h.getSite = getSite[0]
	}
	return h
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	// API 路径不处理
	if strings.HasPrefix(path, "api/") || strings.HasPrefix(path, "d/") {
		http.NotFound(w, r)
		return
	}

	// 尝试查找静态资源（非 index.html 直接走文件服务器）
	if path != "" && path != "index.html" {
		f, err := assets.Open(path)
		if err == nil {
			f.Close()
			h.fileServer.ServeHTTP(w, r)
			return
		}
	}

	// 读取 index.html 并动态注入站点配置
	h.serveIndexHTML(w, r)
}

func (h *SPAHandler) serveIndexHTML(w http.ResponseWriter, r *http.Request) {
	data, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}

	html := string(data)

	// 动态注入站点标题、描述和 logo
	if h.getSite != nil {
		site := h.getSite()
		if site.Title != "" {
			html = strings.Replace(html, "<title>StaticMan</title>", "<title>"+site.Title+"</title>", 1)
			html = strings.Replace(html, `<meta name="description" content="StaticMan`, `<meta name="description" content="`+site.Title, 1)
		}
		if site.Description != "" {
			// 替换 description 内容
			const descPrefix = `<meta name="description" content="`
			if idx := strings.Index(html, descPrefix); idx != -1 {
				start := idx + len(descPrefix)
				if end := strings.Index(html[start:], `"`); end != -1 {
					html = html[:start] + site.Description + html[start+end:]
				}
			}
		}
		if site.Logo != "" {
			html = strings.Replace(html, `<link rel="icon" type="image/svg+xml" href="/logo.svg" />`, `<link rel="icon" type="image/svg+xml" href="`+site.Logo+`" />`, 1)
			html = strings.Replace(html, `<link rel="apple-touch-icon" href="/logo-192.png" />`, `<link rel="apple-touch-icon" href="`+site.Logo+`" />`, 1)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write([]byte(html))
}