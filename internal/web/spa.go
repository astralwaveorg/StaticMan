package web

import (
	"io/fs"
	"net/http"
	"strings"
)

// SiteConfig 用于动态注入 index.html 的站点配置
type SiteConfig struct {
	TitleCN     string
	TitleEN     string
	Title       string // 向后兼容
	Description string
	Logo        string
}

// SiteConfigFunc 获取当前站点配置的回调
type SiteConfigFunc func() SiteConfig

// assets 由 dev.go（开发模式）或 embed.go（withweb 生产模式）初始化
var assets fs.FS

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

		// 浏览器标签页使用中文品牌名
		title := site.TitleCN
		if title == "" {
			title = site.Title
		}
		if title == "" {
			title = "StaticMan"
		}
		html = strings.Replace(html, "<title>StaticMan</title>", "<title>"+title+"</title>", 1)

		// meta description 使用 "英文品牌名 | 描述"
		desc := site.Description
		if site.TitleEN != "" && desc != "" {
			desc = site.TitleEN + " | " + desc
		} else if site.TitleEN != "" {
			desc = site.TitleEN
		} else if desc == "" {
			desc = title
		}
		const descPrefix = `<meta name="description" content="`
		if idx := strings.Index(html, descPrefix); idx != -1 {
			start := idx + len(descPrefix)
			if end := strings.Index(html[start:], `"`); end != -1 {
				html = html[:start] + desc + html[start+end:]
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