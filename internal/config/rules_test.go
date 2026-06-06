package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		pattern  string
		path     string
		expected bool
	}{
		{"*.key", "id_rsa.key", true},
		{"*.key", "id_rsa.pub", false},
		{"ssh/*.key", "ssh/id_rsa.key", true},
		{"**/private/*", "A/B/private/secret.txt", true},
		{"**/private/*", "private/secret.txt", true},
		{"private/", "private/file", true},
		{"regex:.*password.*", "my_password.txt", true},
	}

	for _, tt := range tests {
		expr, _ := globToRegex(tt.pattern)
		re, _ := compileRule(RuleProtect, tt.pattern, "test")
		if re.Regexp.MatchString(tt.path) != tt.expected {
			t.Errorf("pattern %s path %s: expected %v, got %v (regex: %s)", tt.pattern, tt.path, tt.expected, !tt.expected, expr)
		}
	}
}

func TestRuleEngine_Match(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "staticman-test-*")
	defer os.RemoveAll(tempDir)

	engine := NewRuleEngine(tempDir)
	
	// 设置全局规则
	h1, _ := compileRule(RuleHide, ".git/", "global")
	p1, _ := compileRule(RuleProtect, "*.key", "global")
	engine.SetGlobalRules([]Rule{*h1, *p1})

	// 1. 测试全局隐藏
	res := engine.Match(".git/config", true)
	if !res.Hidden {
		t.Errorf("expected .git/config to be hidden")
	}

	// 2. 测试全局保护
	res = engine.Match("id_rsa.key", false)
	if !res.Protected {
		t.Errorf("expected id_rsa.key to be protected")
	}

	// 3. 测试目录级规则覆盖
	os.MkdirAll(filepath.Join(tempDir, "A"), 0755)
	os.WriteFile(filepath.Join(tempDir, "A", ".encrypt"), []byte("hide *.key\n"), 0644)
	
	// 清理缓存以触发重新加载
	engine.ClearCache()
	
	res = engine.Match("A/secret.key", false)
	if !res.Hidden {
		t.Errorf("expected A/secret.key to be hidden by .encrypt")
	}
}
