package masker

import (
	"regexp"
	"strings"
)

// Masker 内容脱敏引擎
type Masker struct {
	patterns []*pattern
}

type pattern struct {
	re       *regexp.Regexp
	valueIdx int // 捕获组中 value 的索引；-1 表示整体替换
}

// New 创建脱敏引擎
func New() *Masker {
	patterns := []*pattern{
		// 明文 key: value (YAML/INI/Conf)
		{re: regexp.MustCompile(`(?i)\b(password|secret|token|api[_-]?key|access[_-]?key|auth|credential|encryption_key)s?\b\s*[:=]\s*["']?([^"'\s,}\)]+)`), valueIdx: 2},
		// JSON "key":"value"
		{re: regexp.MustCompile(`(?i)"(password|secret|token|api[_-]?key|access[_-]?key|auth|credential)s?"\s*:\s*"([^"]+)"`), valueIdx: 2},
		// URI 包含的凭据 (user:pass@) - 整体替换
		{re: regexp.MustCompile(`://[^/\s:]+:[^/\s@]+@`), valueIdx: -1},
		// YAML 中 'user:pass' 凭据
		{re: regexp.MustCompile(`['"]([a-zA-Z0-9_]+:[a-zA-Z0-9_]+)['"]`), valueIdx: 1},
	}

	return &Masker{patterns: patterns}
}

// Mask 对内容进行脱敏处理
func (m *Masker) Mask(content string) string {
	result := content
	for _, p := range m.patterns {
		result = p.re.ReplaceAllStringFunc(result, func(match string) string {
			if p.valueIdx < 0 {
				return "***"
			}
			loc := p.re.FindStringSubmatchIndex(match)
			if loc == nil {
				return "***"
			}
			// valueIdx=2 表示 capture group 2
			// capture group 索引对应的子匹配下标 = 2*valueIdx, 2*valueIdx+1
			valStart := loc[2*p.valueIdx]
			valEnd := loc[2*p.valueIdx+1]
			if valStart < 0 || valEnd < 0 {
				return "***"
			}
			return match[:valStart] + "***" + match[valEnd:]
		})
	}
	return result
}

// MaskLine 对单行进行脱敏处理
func (m *Masker) MaskLine(line string) string {
	return m.Mask(line)
}

// MaskAll 隐藏敏感行
func (m *Masker) MaskAll(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = m.Mask(line)
	}
	return strings.Join(lines, "\n")
}
