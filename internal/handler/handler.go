package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/athena/staticman/internal/auth"
	"github.com/athena/staticman/internal/cache"
	"github.com/athena/staticman/internal/config"
	"github.com/athena/staticman/internal/masker"
	"github.com/athena/staticman/internal/mime"
)

// Handler HTTP 请求处理器
type Handler struct {
	cfg     *config.Config
	masker  *masker.Masker
	dataDir string
	auth    *auth.Service
	cache   *cache.Cache
}

// New 创建新的 Handler
func New(cfg *config.Config, authSvc *auth.Service, c *cache.Cache) *Handler {
	return &Handler{
		cfg:     cfg,
		masker:  masker.New(),
		dataDir: cfg.ConfigsDir(),
		auth:    authSvc,
		cache:   c,
	}
}

// RegisterRoutes 注册所有路由
// 路由架构三层设计:
//
//	1. API 层: /api/* — Web UI 专用，JSON 响应，带认证和脱敏
//	2. 原始文件层: /raw/* 和 /{category}/* — 短 URL，public 文件直接访问，protected 文件需要 ?key=JWT
//	3. 兼容层: /d/* — 旧 URL 重写到新路径
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// === API 层 — Web UI 专用 ===
	mux.HandleFunc("/api/tree", h.handleTree)
	mux.HandleFunc("/api/categories", h.handleCategories)
	mux.HandleFunc("/api/ls", h.handleLs)
	mux.HandleFunc("/api/breadcrumbs", h.handleBreadcrumbs)
	mux.HandleFunc("/api/file/", h.handleFile)
	mux.HandleFunc("/api/search", h.handleSearch)
	mux.HandleFunc("/api/auth", h.handleAuth)
	mux.HandleFunc("/api/config", h.handleConfig)
	mux.HandleFunc("/api/health", h.handleHealth)

	// === 兼容层 — 旧 URL 重写 ===
	mux.HandleFunc("/d/surge/", h.handleLegacySurge)
	mux.HandleFunc("/d/clash/", h.handleLegacyClash)

	// === 原始文件层 — /raw/* 统一 raw URL（支持任意深度） ===
	mux.HandleFunc("/raw/", h.handleRaw)

	// === 原始文件层 — 短 URL 直接映射 ===
	// 扫描 data/ 下所有顶级目录作为分类前缀（排除系统文件）
	entries, err := os.ReadDir(h.dataDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !config.IsSystemFile(entry.Name()) {
				prefix := "/" + entry.Name() + "/"
				mux.HandleFunc(prefix, h.handleRaw)
			}
		}
	}
}

// handleConfig 站点配置
func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	site := h.cfg.GetSite()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"title":       site.Title,
		"description": site.Description,
	})
}

// handleHealth 健康检查
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	total, expired := h.cache.Stats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"version":   "v2",
		"dataDir":   h.dataDir,
		"timestamp": time.Now().Unix(),
		"cache": map[string]interface{}{
			"total":   total,
			"expired": expired,
		},
	})
}

// isAuthenticated 检查请求是否已认证
// 支持三种方式：Authorization header、?key= 查询参数、固定 static_key
func (h *Handler) isAuthenticated(r *http.Request) bool {
	// 1. 检查 Authorization header
	tokenStr := ""
	if auth := r.Header.Get("Authorization"); auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
		tokenStr = auth[7:]
	}
	// 2. 检查 ?key= 查询参数
	if tokenStr == "" {
		tokenStr = r.URL.Query().Get("key")
	}
	if tokenStr == "" {
		return false
	}

	// 3. 优先匹配固定 static_key（简洁、可预测）
	staticKey := h.cfg.GetPassword().StaticKey
	if staticKey != "" && tokenStr == staticKey {
		return true
	}

	// 4. 回退到 JWT 验证
	valid, _ := h.auth.ValidateToken(tokenStr)
	return valid
}

// TreeNode 文件树节点
type TreeNode struct {
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	Type      string     `json:"type"` // "file" or "directory"
	Protected bool       `json:"protected,omitempty"`
	Size      int64      `json:"size,omitempty"`
	ModTime   string     `json:"modTime,omitempty"`
	IsBinary  bool       `json:"isBinary,omitempty"`
	Children  []TreeNode `json:"children,omitempty"`
}

// handleTree 处理文件树请求
// 支持 ?maxDepth=N 限制深度（默认 3，避免过深）
func (h *Handler) handleTree(w http.ResponseWriter, r *http.Request) {
	isAuth := h.isAuthenticated(r)
	maxDepth := 3
	if d := r.URL.Query().Get("maxDepth"); d != "" {
		fmt.Sscanf(d, "%d", &maxDepth)
		if maxDepth < 1 {
			maxDepth = 3
		}
	}
	// 缓存键包含认证状态和深度
	cacheKey := fmt.Sprintf("tree:%v:%d", isAuth, maxDepth)
	if val, ok := h.cache.Get(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(val)
		return
	}
	tree, err := h.buildTree(h.dataDir, "", isAuth, maxDepth, 0)
	if err != nil {
		http.Error(w, "Failed to build tree", http.StatusInternalServerError)
		return
	}
	h.cache.Set(cacheKey, tree, 30*time.Second)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
}

// buildTree 递归构建文件树，可控深度
func (h *Handler) buildTree(rootDir, relPath string, isAuth bool, maxDepth, currentDepth int) (*TreeNode, error) {
	fullPath := filepath.Join(rootDir, relPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	// 根节点特殊处理：使用"root"作为展示名
	name := info.Name()
	if relPath == "" {
		name = "root"
	}

	node := &TreeNode{
		Name:    name,
		Path:    relPath,
		Type:    "file",
		Size:    info.Size(),
		ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
	}

	if info.IsDir() {
		node.Type = "directory"

		// 标记受保护目录
		if h.cfg.IsProtected(relPath) {
			node.Protected = true
		}

		// 超过最大深度，不再递归
		if currentDepth >= maxDepth {
			return node, nil
		}

		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return nil, err
		}

		var children []TreeNode
		for _, entry := range entries {
			childPath := filepath.Join(relPath, entry.Name())
			if relPath == "" {
				childPath = entry.Name()
			}
			child, err := h.buildTree(rootDir, childPath, isAuth, maxDepth, currentDepth+1)
			if err != nil || child == nil {
				continue
			}
			children = append(children, *child)
		}

		sort.Slice(children, func(i, j int) bool {
			if children[i].Type != children[j].Type {
				return children[i].Type == "directory"
			}
			return children[i].Name < children[j].Name
		})

		node.Children = children
	} else {
		// 标记受保护文件
		if h.cfg.IsProtected(relPath) {
			node.Protected = true
		}
		// 标记二进制文件
		if h.isBinaryFile(relPath) {
			node.IsBinary = true
		}
	}

	return node, nil
}

// LsItem 单个 ls 项
type LsItem struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Type      string `json:"type"`
	Size      int64  `json:"size"`
	ModTime   string `json:"modTime"`
	Protected bool   `json:"protected"`
	IsBinary  bool   `json:"isBinary"`
	Language  string `json:"language,omitempty"`
}

// handleLs 列出指定目录的内容（不带 children，用于面包屑点击进入深层）
func (h *Handler) handleLs(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = ""
	}
	// 缓存
	cacheKey := "ls:" + path
	if val, ok := h.cache.Get(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(val)
		return
	}
	fullPath, ok := safeJoin(h.dataDir, path)
	if !ok {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}
	if !info.IsDir() {
		http.Error(w, "Not a directory", http.StatusBadRequest)
		return
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	items := []LsItem{}
	for _, entry := range entries {
		if config.IsSystemFile(entry.Name()) {
			continue
		}
		childPath := filepath.Join(path, entry.Name())
		if path == "" {
			childPath = entry.Name()
		}
		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}
		item := LsItem{
			Name:    entry.Name(),
			Path:    childPath,
			Size:    entryInfo.Size(),
			ModTime: entryInfo.ModTime().Format("2006-01-02 15:04:05"),
		}
		if entry.IsDir() {
			item.Type = "directory"
		} else {
			item.Type = "file"
			item.Language = h.detectLanguage(childPath)
			item.IsBinary = h.isBinaryFile(childPath)
		}
		if h.cfg.IsProtected(childPath) {
			item.Protected = true
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == "directory"
		}
		return items[i].Name < items[j].Name
	})

	result := map[string]interface{}{
		"path":  path,
		"items": items,
		"total": len(items),
	}
	h.cache.Set(cacheKey, result, 10*time.Second)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Breadcrumb 面包屑
type Breadcrumb struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// handleBreadcrumbs 返回路径的面包屑
func (h *Handler) handleBreadcrumbs(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	var crumbs []Breadcrumb

	// 第一级：根
	crumbs = append(crumbs, Breadcrumb{Name: "root", Path: ""})

	if path != "" {
		parts := strings.Split(path, "/")
		current := ""
		for _, part := range parts {
			if current == "" {
				current = part
			} else {
				current = current + "/" + part
			}
			// 第一级目录用 prettifyName 展示，其余用原名
			name := part
			if current == part {
				name = prettifyName(part)
			}
			crumbs = append(crumbs, Breadcrumb{Name: name, Path: current})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(crumbs)
}

// handleCategories 处理分类列表请求
// 完全从 data/ 目录自动扫描生成，无需 metadata.yaml
// 自动推断规则：
//   - name:        目录原名（或首字母大写）
//   - icon:        根据目录名关键词匹配 (proxy→globe, vim→edit-pen, git→branch, shell→terminal, ...)
//   - color:       根据目录名哈希生成稳定的色彩
//   - description: 子工具统计 "包含 N 个工具：..."
func (h *Handler) handleCategories(w http.ResponseWriter, r *http.Request) {
	// 缓存键
	const cacheKey = "categories"
	if val, ok := h.cache.Get(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(val)
		return
	}

	type categoryResult struct {
		Key         string   `json:"key"`
		Name        string   `json:"name"`
		Icon        string   `json:"icon"`
		Description string   `json:"description"`
		Color       string   `json:"color"`
		FileCount   int      `json:"fileCount"`
		Size        int64    `json:"size"`
		Tools       []string `json:"tools"`
	}

	var result []categoryResult

	entries, err := os.ReadDir(h.dataDir)
	if err != nil {
		http.Error(w, "Failed to read data dir", http.StatusInternalServerError)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || config.IsSystemFile(entry.Name()) {
			continue
		}
		key := entry.Name()
		catPath := filepath.Join(h.dataDir, key)

		fileCount, totalSize := h.countFilesAndSize(catPath)
		tools := h.listSubDirs(catPath)

		result = append(result, categoryResult{
			Key:         key,
			Name:        prettifyName(key),
			Icon:        inferIcon(key),
			Description: buildDescription(key, tools, fileCount),
			Color:       inferColor(key),
			FileCount:   fileCount,
			Size:        totalSize,
			Tools:       tools,
		})
	}

	h.cache.Set(cacheKey, result, 30*time.Second)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// listSubDirs 列出直接子目录名
func (h *Handler) listSubDirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs)
	return dirs
}

// prettifyName 美化目录名
// 短横线、下划线 → 空格；只把首字母大写（针对纯 ASCII 单词），中文保持原样
func prettifyName(name string) string {
	// 如果包含中文字符，直接返回（保持用户原意）
	for _, r := range name {
		if r >= 0x4E00 && r <= 0x9FFF {
			return name
		}
	}
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	// 全部小写
	name = strings.ToLower(name)
	// 首字母大写
	if len(name) > 0 {
		return strings.ToUpper(name[:1]) + name[1:]
	}
	return name
}

// inferIcon 根据目录名推断图标
func inferIcon(name string) string {
	n := strings.ToLower(name)
	iconMap := map[string]string{
		"proxy":    "globe",
		"代理":      "globe",
		"代理配置":   "globe",
		"surge":    "shield",
		"clash":    "shield",
		"mihomo":   "shield",
		"vim":      "edit-pen",
		"nvim":     "edit-pen",
		"neovim":   "edit-pen",
		"git":      "branch",
		"shell":    "terminal",
		"bash":     "terminal",
		"zsh":      "terminal",
		"fish":     "terminal",
		"ssh":      "key",
		"docker":   "box",
		"k8s":      "box",
		"kubernetes": "box",
		"script":   "code",
		"scripts":  "code",
		"code":     "code",
		"snippets": "code",
		"note":     "book",
		"notes":    "book",
		"docs":     "book",
		"doc":      "book",
		"fonts":    "type",
		"icon":     "image",
		"icons":    "image",
		"image":    "image",
		"images":   "image",
		"media":    "image",
		"bookmark": "bookmark",
		"bookmarks": "bookmark",
	}
	for k, v := range iconMap {
		if strings.Contains(n, k) {
			return v
		}
	}
	return "folder"
}

// inferColor 根据目录名哈希生成稳定颜色
func inferColor(name string) string {
	h := 0
	for _, r := range name {
		h = h*31 + int(r)
	}
	if h < 0 {
		h = -h
	}
	// 在饱和度/亮度合理的 HSL 空间取色
	hue := h % 360
	colors := []string{
		"#6366f1", "#8b5cf6", "#ec4899", "#f43f5e", "#f97316",
		"#eab308", "#84cc16", "#22c55e", "#10b981", "#14b8a6",
		"#06b6d4", "#0ea5e9", "#3b82f6", "#6366f1", "#a855f7",
	}
	return colors[hue%len(colors)]
}

// buildDescription 自动生成描述
func buildDescription(name string, tools []string, fileCount int) string {
	prettyName := prettifyName(name)
	if len(tools) > 0 {
		prettyTools := make([]string, len(tools))
		for i, t := range tools {
			prettyTools[i] = prettifyName(t)
		}
		return fmt.Sprintf("%s · %d 个工具：%s", prettyName, len(tools), strings.Join(prettyTools, " · "))
	}
	if fileCount == 0 {
		return fmt.Sprintf("%s 目录", prettyName)
	}
	return fmt.Sprintf("%s · %d 个文件", prettyName, fileCount)
}

// countFilesAndSize 递归统计文件数和总大小（跳过系统文件）
func (h *Handler) countFilesAndSize(dir string) (int, int64) {
	count := 0
	var size int64
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0
	}
	for _, entry := range entries {
		if config.IsSystemFile(entry.Name()) {
			continue
		}
		fullPath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			c, s := h.countFilesAndSize(fullPath)
			count += c
			size += s
		} else {
			count++
			info, _ := entry.Info()
			if info != nil {
				size += info.Size()
			}
		}
	}
	return count, size
}

// handleFile 处理文件内容请求（JSON，带脱敏）
func (h *Handler) handleFile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/file/")
	fullPath, ok := safeJoin(h.dataDir, path)
	if !ok {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	isAuth := h.isAuthenticated(r)
	isProtected := h.cfg.IsProtected(path)

	// 对受保护文件拒绝未认证访问（与 /raw/ 行为一致）
	if isProtected && !isAuth {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "禁止访问：此文件需要认证。请在 URL 中添加 ?key=正确的访问秘钥", http.StatusForbidden)
		return
	}

	isBinary := h.isBinaryFile(path)
	language := h.detectLanguage(path)

	// 目录不返回 content
	if info.IsDir() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":        info.Name(),
			"path":        path,
			"type":        "directory",
			"protected":   isProtected,
			"size":        info.Size(),
			"modTime":     info.ModTime().Format("2006-01-02 15:04:05"),
			"isBinary":    false,
			"language":    "",
			"content":     "",
			"truncated":   false,
			"description": h.getFileDescription(path),
		})
		return
	}

	// 二进制文件不返回 content（前端直接显示"无法预览"）
	if isBinary {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":        info.Name(),
			"path":        path,
			"type":        "file",
			"protected":   isProtected,
			"size":        info.Size(),
			"modTime":     info.ModTime().Format("2006-01-02 15:04:05"),
			"isBinary":    true,
			"language":    "",
			"content":     "",
			"truncated":   false,
			"description": h.getFileDescription(path),
		})
		return
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	content := string(data)
	if isProtected && !isAuth {
		content = h.masker.Mask(content)
	}

	// 检查文件大小，超过 1000 行只返回前 1000 行
	truncated := false
	lines := strings.Split(content, "\n")
	if len(lines) > 1000 {
		content = strings.Join(lines[:1000], "\n")
		truncated = true
	}

	result := map[string]interface{}{
		"name":        info.Name(),
		"path":        path,
		"type":        "file",
		"content":     content,
		"language":    language,
		"protected":   isProtected,
		"size":        info.Size(),
		"truncated":   truncated,
		"isBinary":    false,
		"modTime":     info.ModTime().Format("2006-01-02 15:04:05"),
		"description": h.getFileDescription(path),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// getFileDescription 获取文件描述
func (h *Handler) getFileDescription(path string) string {
	meta := h.cfg.GetFileMeta(path)
	if meta != nil {
		return meta.Description
	}
	return ""
}

// handleRaw 通用原始文件服务
// 路径: /raw/<path> 或 /<category>/<path>
func (h *Handler) handleRaw(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	// 优先匹配 /raw/ 前缀
	if strings.HasPrefix(path, "raw/") {
		path = strings.TrimPrefix(path, "raw/")
	} else {
		// 兼容 /<category>/ 前缀
		for _, prefix := range h.getRegisteredPrefixes() {
			if strings.HasPrefix(path, prefix+"/") {
				path = strings.TrimPrefix(path, prefix+"/")
				break
			}
		}
	}

	if h.cfg.IsProtected(path) && !h.isAuthenticated(r) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "禁止访问：此文件需要认证。请在 URL 中添加 ?key=正确的访问秘钥", http.StatusForbidden)
		return
	}
	h.serveRawContent(w, r, path)
}

// getRegisteredPrefixes 获取所有已注册的顶层目录前缀
func (h *Handler) getRegisteredPrefixes() []string {
	entries, err := os.ReadDir(h.dataDir)
	if err != nil {
		return nil
	}
	var prefixes []string
	for _, entry := range entries {
		if entry.IsDir() {
			prefixes = append(prefixes, entry.Name())
		}
	}
	return prefixes
}

// serveRawContent 提供原始文件内容（无脱敏）
func (h *Handler) serveRawContent(w http.ResponseWriter, r *http.Request, path string) {
	fullPath, ok := safeJoin(h.dataDir, path)
	if !ok {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 浏览器预览：对文本类型文件设置 Content-Type 并禁止下载
	contentType := h.detectContentType(path)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	// 对文本类型允许浏览器直接预览
	if isTextContent(contentType) {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(path)))
	}
	w.Write(data)
}

// isTextContent 判断是否为浏览器可预览的文本内容
func isTextContent(contentType string) bool {
	textTypes := []string{"text/", "application/json", "application/xml", "application/javascript"}
	for _, t := range textTypes {
		if strings.HasPrefix(contentType, t) {
			return true
		}
	}
	return false
}

// SearchResult 搜索结果
type SearchResult struct {
	Path      string      `json:"path"`
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	Protected bool        `json:"protected"`
	IsBinary  bool        `json:"isBinary"`
	Language  string      `json:"language,omitempty"`
	Size      int64       `json:"size"`
	Matches   []MatchLine `json:"matches,omitempty"`
}

// MatchLine 匹配行
type MatchLine struct {
	Line int    `json:"line"`
	Text string `json:"text"`
}

// handleSearch 处理搜索请求
func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	searchType := r.URL.Query().Get("type")
	if searchType == "" {
		searchType = "name"
	}

	// 分页参数
	offset := 0
	limit := 50
	if o := r.URL.Query().Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
		if offset < 0 {
			offset = 0
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
		if limit < 1 {
			limit = 50
		}
		if limit > 200 {
			limit = 200
		}
	}

	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []SearchResult{},
			"total":   0,
			"offset":  offset,
			"limit":   limit,
		})
		return
	}

	isAuth := h.isAuthenticated(r)
	var allResults []SearchResult

	if searchType == "name" {
		allResults = h.searchByName(query, isAuth)
	} else {
		allResults = h.searchByContent(query, isAuth)
	}

	total := len(allResults)
	// 分页切片
	end := offset + limit
	if end > total {
		end = total
	}
	var results []SearchResult
	if offset < total {
		results = allResults[offset:end]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"total":   total,
		"offset":  offset,
		"limit":   limit,
	})
}

// searchByName 按文件名搜索
func (h *Handler) searchByName(query string, isAuth bool) []SearchResult {
	var results []SearchResult
	h.walkDir(h.dataDir, "", func(relPath string, info os.FileInfo) {
		if config.IsSystemFile(info.Name()) {
			return
		}
		if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(query)) {
			result := SearchResult{
				Path:      relPath,
				Name:      info.Name(),
				Type:      h.fileType(info),
				Protected: h.cfg.IsProtected(relPath),
				Size:      info.Size(),
			}
			if !info.IsDir() {
				result.IsBinary = h.isBinaryFile(relPath)
				result.Language = h.detectLanguage(relPath)
			}
			results = append(results, result)
		}
	})
	return results
}

// searchByContent 按文件内容搜索
func (h *Handler) searchByContent(query string, isAuth bool) []SearchResult {
	var results []SearchResult
	h.walkDir(h.dataDir, "", func(relPath string, info os.FileInfo) {
		if config.IsSystemFile(info.Name()) {
			return
		}
		if info.IsDir() {
			return
		}
		if h.isBinaryFile(relPath) {
			return
		}

		fullPath := filepath.Join(h.dataDir, relPath)
		// 限制单个文件最大读取 500KB
		f, err := os.Open(fullPath)
		if err != nil {
			return
		}
		defer f.Close()
		data := make([]byte, 500*1024)
		n, _ := f.Read(data)
		data = data[:n]

		content := string(data)
		var matches []MatchLine
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
				matchText := line
				if h.cfg.IsProtected(relPath) && !isAuth {
					matchText = h.masker.Mask(line)
				}
				if len(matchText) > 200 {
					matchText = matchText[:200] + "..."
				}
				matches = append(matches, MatchLine{
					Line: i + 1,
					Text: matchText,
				})
				if len(matches) >= 10 {
					break
				}
			}
		}

		if len(matches) > 0 {
			results = append(results, SearchResult{
				Path:      relPath,
				Name:      info.Name(),
				Type:      "file",
				Protected: h.cfg.IsProtected(relPath),
				IsBinary:  h.isBinaryFile(relPath),
				Language:  h.detectLanguage(relPath),
				Size:      info.Size(),
				Matches:   matches,
			})
		}
	})
	return results
}

// handleAuth 处理认证请求
func (h *Handler) handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	currentPassword := h.cfg.GetPassword().Password
	if !h.auth.VerifyPassword(req.Password, currentPassword) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid password"})
		return
	}

	token, err := h.auth.GenerateToken()
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// walkDir 递归遍历目录
func (h *Handler) walkDir(root, relPath string, fn func(string, os.FileInfo)) {
	fullPath := filepath.Join(root, relPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return
	}
	for _, entry := range entries {
		childPath := filepath.Join(relPath, entry.Name())
		if relPath == "" {
			childPath = entry.Name()
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fn(childPath, info)
		if entry.IsDir() {
			h.walkDir(root, childPath, fn)
		}
	}
}

// detectLanguage 根据文件扩展名判断语法高亮语言
func (h *Handler) detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	langMap := map[string]string{
		".yaml": "yaml", ".yml": "yaml",
		".ini": "ini", ".conf": "ini", ".list": "ini",
		".json": "json",
		".sh": "bash", ".bash": "bash", ".zsh": "bash",
		".vim": "vim", ".vimrc": "vim",
		".toml": "toml",
		".md": "markdown",
		".py": "python",
		".js": "javascript",
		".ts": "typescript",
		".go": "go",
		".css": "css", ".scss": "scss", ".less": "less",
		".html": "xml", ".xml": "xml", ".svg": "xml",
		".sql": "sql",
		".rs": "rust",
		".c": "c", ".h": "c",
		".cpp": "cpp", ".cc": "cpp", ".hpp": "cpp",
		".java": "java",
		".rb": "ruby",
		".php": "php",
		".lua": "lua",
		".log": "plaintext",
		".txt": "plaintext",
		".env": "plaintext",
		"": "plaintext",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	base := filepath.Base(path)
	nameMap := map[string]string{
		"vimrc":     "vim",
		"gitconfig": "ini",
		"zshrc":     "bash",
		"bashrc":    "bash",
		"profile":   "bash",
		"env":       "plaintext",
	}
	if lang, ok := nameMap[base]; ok {
		return lang
	}
	if ext == "" {
		return "plaintext"
	}
	return ""
}

// safeJoin 安全拼接路径
func safeJoin(baseDir, relPath string) (string, bool) {
	cleanRel := filepath.Clean("/" + relPath)
	fullPath := filepath.Join(baseDir, cleanRel)

	rel, err := filepath.Rel(baseDir, fullPath)
	if err != nil {
		return "", false
	}
	if rel == ".." || strings.HasPrefix(rel, "..") {
		return "", false
	}
	return fullPath, true
}

func (h *Handler) detectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	mimeMap := map[string]string{
		".yaml": "text/yaml", ".yml": "text/yaml",
		".ini":  "text/plain",
		".conf": "text/plain",
		".list": "text/plain",
		".md":   "text/markdown",
		".sh":   "text/x-shellscript",
		".py":   "text/x-python",
		".go":   "text/x-go",
		".json": "application/json",
		".js":   "text/javascript",
		".ts":   "text/typescript",
		".css":  "text/css",
		".html": "text/html",
		".xml":  "text/xml",
		".svg":  "image/svg+xml",
		".txt":  "text/plain",
		".log":  "text/plain",
		".png":  "image/png",
		".jpg":  "image/jpeg", ".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".ico":  "image/x-icon",
	}
	if ct, ok := mimeMap[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

func (h *Handler) fileType(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	return "file"
}

// isBinaryFile 判断是否为二进制文件（扩展名 + 魔数双重检测）
func (h *Handler) isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	// 已知文本扩展名 — 快速短路返回
	textExts := map[string]bool{
		".yaml": true, ".yml": true, ".ini": true, ".conf": true,
		".list": true, ".md": true, ".markdown": true,
		".sh": true, ".bash": true, ".zsh": true,
		".vim": true, ".vimrc": true,
		".toml": true, ".json": true, ".xml": true,
		".py": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
		".go": true, ".rs": true, ".c": true, ".h": true, ".cpp": true, ".cc": true,
		".hpp": true, ".java": true, ".rb": true, ".php": true, ".lua": true,
		".css": true, ".scss": true, ".less": true,
		".html": true, ".htm": true, ".svg": true,
		".sql": true, ".txt": true, ".log": true,
		".env": true, ".gitignore": true, ".dockerfile": true,
		".csv": true, ".tsv": true,
	}
	if textExts[ext] {
		return false
	}
	// 已知二进制扩展名 — 快速短路返回
	binaryExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".ico": true, ".webp": true, ".bmp": true,
		".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".7z": true, ".rar": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".db": true, ".sqlite": true, ".sqlite3": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".mp3": true, ".mp4": true, ".wav": true, ".flac": true, ".mov": true, ".avi": true,
		".woff": true, ".woff2": true, ".ttf": true, ".otf": true, ".eot": true,
		".bin": true, ".dat": true,
	}
	if binaryExts[ext] {
		return true
	}
	// 无扩展名或未知扩展名：使用魔数检测
	fullPath := filepath.Join(h.dataDir, path)
	isBin, err := mime.IsBinaryByMagic(fullPath)
	if err == nil {
		return isBin
	}
	// 检测失败时，无扩展名默认视为文本，其他扩展名默认视为二进制（保守策略）
	return ext != ""
}

// handleLegacySurge 兼容旧 Surge URL
func (h *Handler) handleLegacySurge(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/d/surge/")
	newPath := "Surge/" + path
	h.serveRawContent(w, r, newPath)
}

// handleLegacyClash 兼容旧 Clash URL
func (h *Handler) handleLegacyClash(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/d/clash/")
	var newPath string
	switch path {
	case "config.yaml":
		newPath = "Clash/config.yaml"
	case "mihomo/config.yaml":
		newPath = "Clash/config-nas.yaml"
	case "mihomo/config-android.yaml":
		newPath = "Clash/config-android.yaml"
	case "list.yaml":
		newPath = "Clash/nodes.yaml"
	default:
		newPath = "Clash/" + path
	}
	h.serveRawContent(w, r, newPath)
}
