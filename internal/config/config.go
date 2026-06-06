package config

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// PasswordConfig 密码保护配置
type PasswordConfig struct {
	Password  string          `yaml:"password"`
	StaticKey string          `yaml:"static_key"`
	Protected []ProtectedPath `yaml:"protected"`
}

// ProtectedPath 受保护路径规则
type ProtectedPath struct {
	Path string `yaml:"path"` // 相对于 configs/ 的路径
}

// MetadataConfig 文件元数据配置
type MetadataConfig struct {
	Categories map[string]CategoryMeta `yaml:"categories"`
	Files      map[string]FileMeta    `yaml:"files"`
}

// CategoryMeta 分类元数据
type CategoryMeta struct {
	Name        string `yaml:"name"`
	Icon        string `yaml:"icon"`
	Description string `yaml:"description"`
	Color       string `yaml:"color"`
}

// FileMeta 文件元数据
type FileMeta struct {
	Visibility  string   `yaml:"visibility"` // public | protected
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Highlight   string   `yaml:"highlight"` // 语法高亮语言
}

// SiteConfig 站点展示配置
type SiteConfig struct {
	TitleCN     string `yaml:"title_cn"`
	TitleEN     string `yaml:"title_en"`
	Title       string `yaml:"title"` // 向后兼容：完整标题
	Description string `yaml:"description"`
	Logo        string `yaml:"logo"`
}

// Config 应用配置
type Config struct {
	mu            sync.RWMutex
	DataDir       string
	Password      PasswordConfig
	Metadata      MetadataConfig
	Site          SiteConfig
	AccessKeyHash string // JWT 签名密钥

	passwordModTime time.Time
	metadataModTime time.Time
}

// Load 从数据目录加载配置
func Load(dataDir string) (*Config, error) {
	c := &Config{DataDir: dataDir}

	if err := c.loadPassword(); err != nil {
		return nil, err
	}
	if err := c.loadMetadata(); err != nil {
		return nil, err
	}

	// JWT 密钥：优先环境变量，否则用密码
	c.AccessKeyHash = os.Getenv("ACCESS_KEY")
	if c.AccessKeyHash == "" {
		c.AccessKeyHash = c.Password.Password
	}

	// 站点标题和描述：优先环境变量，默认使用项目名
	siteTitle := firstNonEmpty(os.Getenv("SITE_TITLE"), "StaticMan")
	siteDesc := firstNonEmpty(os.Getenv("SITE_DESCRIPTION"), "StaticMan - 私人网络配置管理中心")
	siteLogo := firstNonEmpty(os.Getenv("SITE_LOGO"), "/logo.svg")

	// 中英文品牌名支持：可独立配置，未配置时从 SITE_TITLE 回退
	siteTitleCN := firstNonEmpty(os.Getenv("SITE_TITLE_CN"), extractChinese(siteTitle), siteTitle)
	siteTitleEN := firstNonEmpty(os.Getenv("SITE_TITLE_EN"), extractEnglish(siteTitle), siteTitle)

	c.Site = SiteConfig{
		TitleCN:     siteTitleCN,
		TitleEN:     siteTitleEN,
		Title:       siteTitle,
		Description: siteDesc,
		Logo:        siteLogo,
	}

	return c, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// extractChinese 从字符串中提取连续的中文字符（取最长的一段）
func extractChinese(s string) string {
	var buf []rune
	var best []rune
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			buf = append(buf, r)
		} else {
			if len(buf) > len(best) {
				best = buf
			}
			buf = nil
		}
	}
	if len(buf) > len(best) {
		best = buf
	}
	return string(best)
}

// extractEnglish 从字符串中提取连续的 ASCII 字母/数字（取最长的一段）
func extractEnglish(s string) string {
	var buf []rune
	var best []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			buf = append(buf, r)
		} else {
			if len(buf) > len(best) {
				best = buf
			}
			buf = nil
		}
	}
	if len(buf) > len(best) {
		best = buf
	}
	return string(best)
}

// Watch 启动配置文件热加载
func (c *Config) Watch() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			c.reloadIfNeeded()
		}
	}()
}

func (c *Config) loadPassword() error {
	path := filepath.Join(c.DataDir, "password.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.Password = PasswordConfig{Password: "", Protected: nil}
			return nil
		}
		return err
	}
	if err := yaml.Unmarshal(data, &c.Password); err != nil {
		return err
	}
	info, _ := os.Stat(path)
	if info != nil {
		c.passwordModTime = info.ModTime()
	}
	return nil
}

func (c *Config) loadMetadata() error {
	path := filepath.Join(c.DataDir, "metadata.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.Metadata = MetadataConfig{Categories: map[string]CategoryMeta{}, Files: map[string]FileMeta{}}
			return nil
		}
		return err
	}
	if err := yaml.Unmarshal(data, &c.Metadata); err != nil {
		return err
	}
	info, _ := os.Stat(path)
	if info != nil {
		c.metadataModTime = info.ModTime()
	}
	return nil
}

func (c *Config) reloadIfNeeded() {
	c.mu.Lock()
	defer c.mu.Unlock()

	passwordPath := filepath.Join(c.DataDir, "password.yaml")
	metadataPath := filepath.Join(c.DataDir, "metadata.yaml")

	if info, err := os.Stat(passwordPath); err == nil && info.ModTime().After(c.passwordModTime) {
		c.loadPassword()
		c.passwordModTime = info.ModTime()
	}

	if info, err := os.Stat(metadataPath); err == nil && info.ModTime().After(c.metadataModTime) {
		c.loadMetadata()
		c.metadataModTime = info.ModTime()
	}
}

// GetPassword 返回当前密码配置（线程安全）
func (c *Config) GetPassword() PasswordConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Password
}

// GetMetadata 返回当前元数据配置（线程安全）
func (c *Config) GetMetadata() MetadataConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Metadata
}

// GetSite 返回站点展示配置（线程安全）
func (c *Config) GetSite() SiteConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Site
}

// IsProtected 检查路径是否受保护
// 统一保护模型：protected 表示需要认证，public 表示公开
func (c *Config) IsProtected(path string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 检查 password.yaml 中的 protected 列表
	for _, p := range c.Password.Protected {
		if path == p.Path || isPathUnder(path, p.Path) {
			return true
		}
	}

	// 检查 metadata.yaml 中的可见性
	if meta, ok := c.Metadata.Files[path]; ok {
		if meta.Visibility == "protected" {
			return true
		}
	}

	return false
}

// isPathUnder 检查 path 是否在 prefix 下
func isPathUnder(path, prefix string) bool {
	if len(path) <= len(prefix) {
		return false
	}
	return path[:len(prefix)+1] == prefix+"/"
}

// GetFileMeta 获取文件元数据
func (c *Config) GetFileMeta(path string) *FileMeta {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if meta, ok := c.Metadata.Files[path]; ok {
		return &meta
	}
	return nil
}

// ConfigsDir 返回配置文件目录
// 即 dataDir 自身，但排除系统配置文件（password.yaml 等）
func (c *Config) ConfigsDir() string {
	return c.DataDir
}

// IsSystemFile 判断是否为系统配置文件（非用户内容，不应出现在导航中）
func IsSystemFile(name string) bool {
	systemFiles := map[string]bool{
		"password.yaml": true,
		"password.yml":  true,
		"metadata.yaml": true,
		"metadata.yml":  true,
                ".git":          true,
                ".github":       true,
                ".DS_Store":     true,
	}
	return systemFiles[name]
}