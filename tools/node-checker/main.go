package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Node struct {
	Name    string
	Type    string
	Server  string
	Port    int
	Password string
	Extra   map[string]string
	RawLine string
}

type UnlockResult struct {
	Netflix  string
	YouTube   string
	OpenAI    string
	Disney    string
	Gemini    string
	Claude    string
	TikTok    string
	IP        string
	Country   string
	IPRisk    string
	Speed     int
	Alive     bool
}

var cfg = struct {
	Concurrent   int
	Timeout      int
	MediaTimeout int
	AliveTestURL string
	MediaCheck   bool
	Platforms    string
}{
	Concurrent:   30,
	Timeout:      5000,
	MediaTimeout: 10,
	AliveTestURL: "http://gstatic.com/generate_204",
	MediaCheck:   true,
	Platforms:    "netflix,youtube,openai,disney,gemini,claude,tiktok,iprisk",
}

var progress atomic.Int32
var available atomic.Int32

func main() {
	inputFile := flag.String("i", "", "输入节点文件 (Surge .ini)")
	outputDir := flag.String("o", "", "输出目录")
	concurrent := flag.Int("j", 30, "并发数")
	timeout := flag.Int("t", 5000, "超时时间 (毫秒)")
	noMedia := flag.Bool("no-media", false, "跳过媒体解锁检测")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: node-checker -i <input.ini> [-o <output-dir>] [-j 30] [-t 5000] [--no-media]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg.Concurrent = *concurrent
	cfg.Timeout = *timeout
	if *noMedia {
		cfg.MediaCheck = false
	}

	nodes, err := parseSurgeINI(*inputFile)
	if err != nil {
		slog.Error("解析输入文件失败", "错误", err)
		os.Exit(1)
	}

	slog.Info("成功解析节点", "数量", len(nodes), "文件", *inputFile)

	if len(nodes) == 0 {
		slog.Warn("未找到任何节点")
		os.Exit(0)
	}

	if *outputDir == "" {
		*outputDir = filepath.Dir(*inputFile)
	}

	fmt.Printf("开始检测 %d 个节点，并发: %d，超时: %dms\n", len(nodes), cfg.Concurrent, cfg.Timeout)
	startTime := time.Now()
	results, _ := checkNodes(nodes)
	elapsed := time.Since(startTime)

	slog.Info("检测完成", "耗时", elapsed.Round(time.Second), "可用", available.Load(), "总数", len(nodes))

	writeOutputs(filepath.Join(*outputDir, "all-checked.ini"), results, nodes)
	printSummary(results)
}

func parseSurgeINI(path string) ([]Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var nodes []Node
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		node, err := parseNodeLine(line)
		if err != nil {
			slog.Debug("解析行失败", "行", line, "错误", err)
			continue
		}
		if node != nil {
			nodes = append(nodes, *node)
		}
	}
	return nodes, scanner.Err()
}

func parseNodeLine(line string) (*Node, error) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("no '=' found")
	}

	name := strings.TrimSpace(parts[0])
	rest := strings.TrimSpace(parts[1])

	segments := splitWithQuotes(rest)
	if len(segments) < 3 {
		return nil, fmt.Errorf("字段不足")
	}

	nodeType := strings.TrimSpace(segments[0])
	server := strings.TrimSpace(segments[1])
	portStr := strings.TrimSpace(segments[2])

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("无效端口: %s", portStr)
	}

	node := &Node{
		Name:    name,
		Type:    strings.ToLower(nodeType),
		Server:  server,
		Port:    port,
		RawLine: line,
		Extra:   make(map[string]string),
	}

	for i := 3; i < len(segments); i++ {
		kv := splitKeyValue(segments[i])
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := stripQuotes(strings.TrimSpace(kv[1]))
			switch key {
			case "password", "uuid":
				node.Password = value
			case "encrypt-method", "cipher":
				node.Extra["cipher"] = value
			case "sni", "servername":
				node.Extra["sni"] = value
			case "tls", "security":
				node.Extra["tls"] = value
			default:
				node.Extra[key] = value
			}
		}
	}
	return node, nil
}

func splitWithQuotes(s string) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuote = !inQuote
			current.WriteByte(c)
		} else if c == ',' && !inQuote {
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}

func splitKeyValue(s string) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuote = !inQuote
			current.WriteByte(c)
		} else if c == '=' && !inQuote {
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
		   (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func checkNodes(nodes []Node) ([]UnlockResult, error) {
	results := make([]UnlockResult, len(nodes))
	total := len(nodes)

	progress.Store(0)
	available.Store(0)

	concurrency := cfg.Concurrent
	if concurrency > total {
		concurrency = total
	}

	type checkTask struct {
		idx  int
		node Node
	}

	tasks := make(chan checkTask, concurrency)
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				r := checkSingleNode(task.node)
				results[task.idx] = r
				progress.Add(1)
				if r.Alive {
					available.Add(1)
				}
			}
		}()
	}

	go func() {
		defer close(tasks)
		for i, node := range nodes {
			tasks <- checkTask{idx: i, node: node}
		}
	}()

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p := progress.Load()
				a := available.Load()
				fmt.Printf("\r[%d/%d] 可用: %d", p, total, a)
				if p >= int32(total) {
					return
				}
			}
		}
	}()

	wg.Wait()
	fmt.Println()

	return results, nil
}

func checkSingleNode(node Node) UnlockResult {
	result := UnlockResult{}

	// TCP 端口可达性检测（最可靠的方式）
	if !tcpTest(node.Server, node.Port, cfg.Timeout/1000) {
		slog.Debug("TCP 端口不可达", "节点", node.Name, "服务器", node.Server, "端口", node.Port)
		return result
	}

	slog.Debug("TCP 端口可达", "节点", node.Name, "服务器", node.Server, "端口", node.Port)
	result.Alive = true

	// 对 hysteria2 类型做 HTTP 解锁检测
	if cfg.MediaCheck && (node.Type == "hysteria2" || node.Type == "hy2") {
		checkMediaHysteria2(node, &result)
	}

	return result
}

// tcpTest 用 bash /dev/tcp 测试 TCP 端口可达性
func tcpTest(host string, port int, timeoutSec int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("echo >/dev/tcp/%s/%d", host, port))
	return cmd.Run() == nil
}

// checkMediaHysteria2 用 curl 测试 hysteria2 节点的流媒体解锁
func checkMediaHysteria2(node Node, result *UnlockResult) {
	proxyURL := fmt.Sprintf("http://%s:%d", node.Server, node.Port)
	timeout := fmt.Sprintf("%d", cfg.MediaTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.MediaTimeout)*time.Second)
	defer cancel()

	// Netflix
	if strings.Contains(cfg.Platforms, "netflix") {
		cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			"--proxy", proxyURL,
			"--connect-timeout", timeout,
			"-L", "https://www.netflix.com/title/81280792")
		output, _ := cmd.Output()
		code := strings.TrimSpace(string(output))
		if code == "200" {
			result.Netflix = "NF-US"
		} else if code == "404" {
			result.Netflix = "NF"
		}
	}

	// YouTube
	if strings.Contains(cfg.Platforms, "youtube") {
		cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			"--proxy", proxyURL,
			"--connect-timeout", timeout,
			"https://www.youtube.com/premium")
		output, _ := cmd.Output()
		code := strings.TrimSpace(string(output))
		if code == "200" {
			result.YouTube = "US"
		}
	}

	// OpenAI
	if strings.Contains(cfg.Platforms, "openai") {
		cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			"--proxy", proxyURL,
			"--connect-timeout", timeout,
			"https://api.openai.com/compliance/cookie_requirements")
		output, _ := cmd.Output()
		code := strings.TrimSpace(string(output))
		if code == "200" {
			bodyCmd := exec.CommandContext(ctx, "curl", "-s",
				"--proxy", proxyURL,
				"--connect-timeout", timeout,
				"https://api.openai.com/compliance/cookie_requirements")
			body, _ := bodyCmd.Output()
			if !strings.Contains(strings.ToLower(string(body)), "unsupported_country") {
				result.OpenAI = "GPT⁺"
			} else {
				result.OpenAI = "GPT"
			}
		}
	}

	// Disney+
	if strings.Contains(cfg.Platforms, "disney") {
		cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			"--proxy", proxyURL,
			"--connect-timeout", timeout,
			"https://www.disneyplus.com/")
		output, _ := cmd.Output()
		code := strings.TrimSpace(string(output))
		if code == "200" {
			result.Disney = "D+"
		}
	}

	// Gemini
	if strings.Contains(cfg.Platforms, "gemini") {
		cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			"--proxy", proxyURL,
			"--connect-timeout", timeout,
			"https://gemini.google.com/")
		output, _ := cmd.Output()
		code := strings.TrimSpace(string(output))
		if code == "200" {
			result.Gemini = "GM"
		}
	}
}

func writeOutputs(path string, results []UnlockResult, nodes []Node) {
	file, err := os.Create(path)
	if err != nil {
		slog.Error("创建输出文件失败", "错误", err)
		return
	}
	defer file.Close()

	buf := bufio.NewWriter(file)
	defer buf.Flush()

	header := fmt.Sprintf("# MagicHub 节点检测结果\n# 生成时间: %s\n# 总节点数: %d\n\n",
		time.Now().UTC().Format("2006-01-02 15:04:05 MST"), len(nodes))
	buf.WriteString(header)

	for i, node := range nodes {
		r := results[i]
		if !r.Alive {
			continue
		}
		name := buildNodeName(node.Name, r)
		line := buildSurgeLine(name, node)
		buf.WriteString(line + "\n")
	}
}

func buildNodeName(baseName string, r UnlockResult) string {
	name := baseName
	if r.Netflix != "" {
		name += " [" + r.Netflix + "]"
	}
	if r.YouTube != "" {
		name += " [YT-" + r.YouTube + "]"
	}
	if r.OpenAI != "" {
		name += " [" + r.OpenAI + "]"
	}
	if r.Disney != "" {
		name += " [" + r.Disney + "]"
	}
	if r.Gemini != "" {
		name += " [GM-" + r.Gemini + "]"
	}
	return name
}

func buildSurgeLine(name string, node Node) string {
	switch node.Type {
	case "ss":
		cipher := node.Extra["cipher"]
		if cipher == "" {
			cipher = "aes-256-gcm"
		}
		return fmt.Sprintf("%s = ss, %s, %d, encrypt-method=%s, password=%s",
			name, node.Server, node.Port, cipher, node.Password)
	case "vmess":
		tls := "false"
		if node.Extra["tls"] == "tls" || node.Extra["tls"] == "true" {
			tls = "true"
		}
		return fmt.Sprintf("%s = vmess, %s, %d, username=%s, tls=%s",
			name, node.Server, node.Port, node.Password, tls)
	case "vless":
		tls := "false"
		if node.Extra["tls"] == "tls" || node.Extra["tls"] == "true" {
			tls = "true"
		}
		return fmt.Sprintf("%s = vless, %s, %d, username=%s, tls=%s",
			name, node.Server, node.Port, node.Password, tls)
	case "trojan":
		return fmt.Sprintf("%s = trojan, %s, %d, password=%s",
			name, node.Server, node.Port, node.Password)
	case "hysteria2", "hy2":
		sni := node.Extra["sni"]
		skipCert := node.Extra["skip-cert-verify"]
		line := fmt.Sprintf("%s = hysteria2, %s, %d, password=%s", name, node.Server, node.Port, node.Password)
		if sni != "" {
			line += fmt.Sprintf(", sni=%s", sni)
		}
		if skipCert == "true" {
			line += ", skip-cert-verify=true"
		}
		return line
	default:
		return fmt.Sprintf("%s = %s, %s, %d",
			name, node.Type, node.Server, node.Port)
	}
}

func printSummary(results []UnlockResult) {
	stats := map[string]int{
		"total":   len(results),
		"alive":   0,
		"netflix": 0,
		"youtube": 0,
		"openai":  0,
		"disney":  0,
		"gemini":  0,
	}

	for _, r := range results {
		if !r.Alive {
			continue
		}
		stats["alive"]++
		if r.Netflix != "" {
			stats["netflix"]++
		}
		if r.YouTube != "" {
			stats["youtube"]++
		}
		if r.OpenAI != "" {
			stats["openai"]++
		}
		if r.Disney != "" {
			stats["disney"]++
		}
		if r.Gemini != "" {
			stats["gemini"]++
		}
	}

	fmt.Println("\n=== 检测结果汇总 ===")
	fmt.Printf("总节点数: %d\n", stats["total"])
	fmt.Printf("可用节点: %d (%.1f%%)\n", stats["alive"], float64(stats["alive"])/float64(stats["total"])*100)
	fmt.Println("\n平台解锁统计 (仅 hysteria2 节点):")
	fmt.Printf("  Netflix: %d\n", stats["netflix"])
	fmt.Printf("  YouTube: %d\n", stats["youtube"])
	fmt.Printf("  OpenAI: %d\n", stats["openai"])
	fmt.Printf("  Disney+: %d\n", stats["disney"])
	fmt.Printf("  Gemini: %d\n", stats["gemini"])
}
