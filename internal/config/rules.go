package config

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// RuleType 规则策略类型
type RuleType string

const (
	RuleHide    RuleType = "hide"
	RuleProtect RuleType = "protect"
)

// Rule 单条匹配规则
type Rule struct {
	Type      RuleType       // hide 或 protect
	Pattern   string         // 原始模式字符串
	Regexp    *regexp.Regexp // 预编译的正则表达式
	IsAbs     bool           // 是否为从根开始的绝对路径 (以 / 开头)
	IsDirOnly bool           // 是否仅匹配目录 (以 / 结尾)
	Source    string         // 规则来源：global 或 .encrypt 文件路径
}

// MatchResult 匹配结果
type MatchResult struct {
	Hidden    bool  // 是否隐藏
	Protected bool  // 是否受保护
	MatchedBy *Rule // 命中的规则
}

// RuleEngine 规则引擎
type RuleEngine struct {
	mu          sync.RWMutex
	globalRules []Rule                 // 全局规则 (来自 password.yaml)
	dirRules    map[string][]Rule      // 目录级规则缓存 (key 为目录相对路径)
	cache       map[string]MatchResult // 路径匹配结果缓存
	dataDir     string                 // 数据根目录
}

// NewRuleEngine 创建新的规则引擎
func NewRuleEngine(dataDir string) *RuleEngine {
	return &RuleEngine{
		dirRules: make(map[string][]Rule),
		cache:    make(map[string]MatchResult),
		dataDir:  dataDir,
	}
}

// globToRegex 将 glob 模式转换为正则表达式
func globToRegex(pattern string) (string, bool) {
	if strings.HasPrefix(pattern, "regex:") {
		return strings.TrimPrefix(pattern, "regex:"), true
	}

	var sb strings.Builder
	sb.WriteString("^")

	// 处理 **/ 开头
	if strings.HasPrefix(pattern, "**/") {
		sb.WriteString("(.*)?")
		pattern = pattern[3:]
	} else if !strings.HasPrefix(pattern, "/") {
		// 非绝对路径且非 ** 开头，隐含支持前缀 (如 .git/ 匹配任意层级的 .git)
		sb.WriteString("(.*)?")
	} else {
		// 绝对路径，去掉 /
		pattern = pattern[1:]
	}

	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				sb.WriteString(".*")
				i += 2
			} else {
				sb.WriteString("[^/]*")
				i++
			}
		case '?':
			sb.WriteString("[^/]")
			i++
		case '.', '+', '(', ')', '|', '[', ']', '{', '}', '^', '$', '\\':
			sb.WriteString("\\")
			sb.WriteByte(pattern[i])
			i++
		case '/':
			// 如果是中间的 /，直接写
			// 如果是结尾的 /，支持匹配目录下所有内容
			if i == len(pattern)-1 {
				sb.WriteString("(/.*)?")
			} else {
				sb.WriteByte('/')
			}
			i++
		default:
			sb.WriteByte(pattern[i])
			i++
		}
	}
	
	// 如果模式不以 / 结尾，也要支持匹配子路径（如果是目录规则）
	// 但这里我们简单点：完全匹配模式
	if !strings.HasSuffix(pattern, "/") {
		sb.WriteString("$")
	}
	
	return sb.String(), false
}

// compileRule 将原始规则字符串编译为 Rule 结构体
func compileRule(type_ RuleType, pattern, source string) (*Rule, error) {
	isAbs := strings.HasPrefix(pattern, "/")
	isDirOnly := strings.HasSuffix(pattern, "/")

	expr, _ := globToRegex(pattern)
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	return &Rule{
		Type:      type_,
		Pattern:   pattern,
		Regexp:    re,
		IsAbs:     isAbs,
		IsDirOnly: isDirOnly,
		Source:    source,
	}, nil
}

// Match 对路径进行规则匹配
func (e *RuleEngine) Match(relPath string, isDir bool) MatchResult {
	relPath = strings.ReplaceAll(relPath, "\\", "/")
	relPath = strings.Trim(relPath, "/")

	e.mu.RLock()
	if res, ok := e.cache[relPath]; ok {
		e.mu.RUnlock()
		return res
	}
	e.mu.RUnlock()

	result := MatchResult{}
	var ruleChain [][]Rule

	e.mu.RLock()
	ruleChain = append(ruleChain, e.globalRules)
	e.mu.RUnlock()

	curr := relPath
	if !isDir && relPath != "" {
		curr = filepath.Dir(relPath)
	}
	if curr == "." {
		curr = ""
	}

	for {
		rules := e.getOrLoadDirRules(curr)
		if len(rules) > 0 {
			ruleChain = append([][]Rule{rules}, ruleChain...)
		}
		if curr == "" {
			break
		}
		curr = filepath.Dir(curr)
		if curr == "." {
			curr = ""
		}
	}

	for _, rules := range ruleChain {
		for _, rule := range rules {
			if rule.Type == RuleHide && e.matchRule(&rule, relPath, isDir) {
				result.Hidden = true
				result.MatchedBy = &rule
				goto end
			}
		}
	}

	for _, rules := range ruleChain {
		for _, rule := range rules {
			if rule.Type == RuleProtect && e.matchRule(&rule, relPath, isDir) {
				result.Protected = true
				result.MatchedBy = &rule
				goto end
			}
		}
	}

end:
	e.mu.Lock()
	e.cache[relPath] = result
	e.mu.Unlock()

	return result
}

func (e *RuleEngine) getOrLoadDirRules(dir string) []Rule {
	e.mu.RLock()
	rules, ok := e.dirRules[dir]
	e.mu.RUnlock()
	if ok {
		return rules
	}

	path := filepath.Join(e.dataDir, dir, ".encrypt")
	_, err := os.Stat(path)
	if err != nil {
		e.mu.Lock()
		e.dirRules[dir] = nil
		e.mu.Unlock()
		return nil
	}

	newRules := e.parseEncryptFile(path, dir)
	e.mu.Lock()
	e.dirRules[dir] = newRules
	e.mu.Unlock()
	return newRules
}

func (e *RuleEngine) parseEncryptFile(path, sourceDir string) []Rule {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var rules []Rule
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		type_ := RuleProtect
		pattern := line
		if strings.HasPrefix(line, "hide ") {
			type_ = RuleHide
			pattern = strings.TrimSpace(strings.TrimPrefix(line, "hide "))
		}

		if r, err := compileRule(type_, pattern, path); err == nil {
			rules = append(rules, *r)
		}
	}
	return rules
}

func (e *RuleEngine) ClearCache() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cache = make(map[string]MatchResult)
	e.dirRules = make(map[string][]Rule)
}

func (e *RuleEngine) SetGlobalRules(rules []Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.globalRules = rules
	e.cache = make(map[string]MatchResult)
}

func (e *RuleEngine) matchRule(r *Rule, path string, isDir bool) bool {
	if r.IsDirOnly {
		// 目录规则（如 .git/）匹配目录本身及其下所有内容
		dirPattern := strings.TrimSuffix(r.Pattern, "/")
		if path == dirPattern || strings.HasPrefix(path, dirPattern+"/") {
			return true
		}
		return false
	}
	return r.Regexp.MatchString(path)
}
