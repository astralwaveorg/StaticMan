package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/constant"
)

// Node represents a parsed Surge node
type Node struct {
	Name     string
	Type     string
	Server   string
	Port     int
	Password string
	Extra    map[string]string
	RawLine  string
}

// UnlockResult represents unlock detection results
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
	inputFile := flag.String("i", "", "Input Surge .ini file")
	outputDir := flag.String("o", "", "Output directory")
	configFile := flag.String("c", "", "Config YAML file")
	concurrent := flag.Int("j", 30, "Concurrency")
	timeout := flag.Int("t", 5000, "Timeout in milliseconds")
	noMedia := flag.Bool("no-media", false, "Skip media unlock detection")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: node-checker -i <input.ini> [-o <output-dir>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg.Concurrent = *concurrent
	cfg.Timeout = *timeout
	if *noMedia {
		cfg.MediaCheck = false
	}

	if *configFile != "" {
		parseConfig(*configFile)
	}

	nodes, err := parseSurgeINI(*inputFile)
	if err != nil {
		slog.Error("Failed to parse input file", "error", err)
		os.Exit(1)
	}

	slog.Info("Parsed nodes", "count", len(nodes))

	if len(nodes) == 0 {
		slog.Warn("No nodes found")
		os.Exit(0)
	}

	if *outputDir == "" {
		*outputDir = filepath.Dir(*inputFile)
	}

	startTime := time.Now()
	results, _ := checkNodes(nodes)
	elapsed := time.Since(startTime)

	slog.Info("Check completed", "elapsed", elapsed.Round(time.Second), "available", available.Load(), "total", len(nodes))

	writeOutputs(filepath.Join(*outputDir, "all-checked.ini"), results, nodes)
	printSummary(results)
}

func parseConfig(path string) {
	slog.Info("Config file parsed (placeholder)", "path", path)
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
			slog.Debug("Failed to parse line", "line", line, "error", err)
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

	// Split by comma, but respect quotes
	segments := splitWithQuotes(rest)
	if len(segments) < 3 {
		return nil, fmt.Errorf("not enough segments")
	}

	nodeType := strings.TrimSpace(segments[0])
	server := strings.TrimSpace(segments[1])
	portStr := strings.TrimSpace(segments[2])

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
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
			value := strings.TrimSpace(kv[1])
			switch key {
			case "password", "uuid":
				node.Password = stripQuotes(value)
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

// splitWithQuotes splits a string by comma, respecting quoted strings
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

// splitKeyValue splits a key=value pair, respecting quotes
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

// stripQuotes removes surrounding quotes from a string
func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
		   (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func (n *Node) toMihomoProxy() map[string]any {
	m := make(map[string]any)
	m["name"] = n.Name
	m["type"] = n.Type
	m["server"] = n.Server
	m["port"] = n.Port

	if n.Password != "" {
		if n.Type == "vmess" {
			m["uuid"] = n.Password
		} else {
			m["password"] = n.Password
		}
	}

	// Check for shadow-tls plugin
	hasShadowTLS := false
	pluginOpts := ""

	if sni, ok := n.Extra["shadow-tls-sni"]; ok && sni != "" {
		hasShadowTLS = true
		if pluginOpts != "" {
			pluginOpts += "; "
		}
		pluginOpts += "tls-sni=" + sni
	}
	if pwd, ok := n.Extra["shadow-tls-password"]; ok && pwd != "" {
		hasShadowTLS = true
		if pluginOpts != "" {
			pluginOpts += "; "
		}
		pluginOpts += "password=" + pwd
	}
	if ver, ok := n.Extra["shadow-tls-version"]; ok && ver != "" {
		hasShadowTLS = true
		if pluginOpts != "" {
			pluginOpts += "; "
		}
		pluginOpts += "version=" + ver
	}

	if hasShadowTLS {
		m["plugin"] = "shadow-tls"
		m["plugin-opts"] = pluginOpts
	}

	for k, v := range n.Extra {
		// Skip shadow-tls params as they're handled above
		if strings.HasPrefix(k, "shadow-tls-") {
			continue
		}
		// Skip-cert-verify -> insecure for mihomo
		if k == "skip-cert-verify" && v == "true" {
			m["insecure"] = true
			continue
		}
		m[k] = v
	}
	return m
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
				fmt.Printf("\r[%d/%d] alive: %d", p, total, a)
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

	proxyMap := node.toMihomoProxy()
	proxy, err := adapter.ParseProxy(proxyMap)
	if err != nil {
		slog.Debug("Failed to parse proxy", "name", node.Name, "error", err)
		return result
	}
	defer proxy.Close()

	httpClient, err := createProxyClient(proxy)
	if err != nil {
		slog.Debug("Failed to create proxy client", "name", node.Name, "error", err)
		return result
	}
	defer httpClient.CloseIdleConnections()

	alive, err := checkAlive(httpClient)
	if err != nil || !alive {
		return result
	}
	result.Alive = true

	if cfg.MediaCheck {
		checkMedia(httpClient, &result)
	}

	return result
}

func createProxyClient(proxy constant.Proxy) (*http.Client, error) {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, portStr, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			port, err := strconv.ParseUint(portStr, 10, 16)
			if err != nil {
				return nil, err
			}
			return proxy.DialContext(ctx, &constant.Metadata{
				Host:    host,
				DstPort: uint16(port),
			})
		},
		DisableKeepAlives: true,
	}

	return &http.Client{
		Timeout:   time.Duration(cfg.Timeout) * time.Millisecond,
		Transport: transport,
	}, nil
}

func checkAlive(client *http.Client) (bool, error) {
	req, err := http.NewRequest("GET", cfg.AliveTestURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200 || resp.StatusCode == 204, nil
}

func checkMedia(client *http.Client, result *UnlockResult) {
	if strings.Contains(cfg.Platforms, "netflix") {
		if r := checkNetflix(client); r != "" {
			result.Netflix = r
		}
	}
	if strings.Contains(cfg.Platforms, "youtube") {
		if r := checkYouTube(client); r != "" {
			result.YouTube = r
		}
	}
	if strings.Contains(cfg.Platforms, "openai") {
		if r := checkOpenAI(client); r != "" {
			result.OpenAI = r
		}
	}
	if strings.Contains(cfg.Platforms, "disney") {
		if checkDisney(client) {
			result.Disney = "D+"
		}
	}
	if strings.Contains(cfg.Platforms, "gemini") {
		if checkGemini(client) {
			result.Gemini = "GM"
		}
	}
}

func checkNetflix(client *http.Client) string {
	req, _ := http.NewRequest("GET", "https://www.netflix.com/title/81280792", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 301 {
		region := getNetflixRegion(client)
		return "NF-" + region
	}

	if resp.StatusCode == 404 {
		req2, _ := http.NewRequest("GET", "https://www.netflix.com/title/70143836", nil)
		req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		resp2, err := client.Do(req2)
		if err == nil {
			defer resp2.Body.Close()
			if resp2.StatusCode == 200 {
				return "NF"
			}
		}
	}
	return ""
}

func getNetflixRegion(client *http.Client) string {
	req, _ := http.NewRequest("GET", "https://www.netflix.com/title/80018499", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	noRedirectClient := *client
	noRedirectClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "US"
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		return "US"
	}

	for _, part := range strings.Split(location, "/") {
		if len(part) == 2 && part[0] >= 'A' && part[0] <= 'Z' {
			return strings.ToUpper(part)
		}
	}
	return "US"
}

func checkYouTube(client *http.Client) string {
	req, _ := http.NewRequest("GET", "https://www.youtube.com/premium", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		buf := new(strings.Builder)
		io.Copy(buf, resp.Body)
		body := buf.String()

		if strings.Contains(body, "Premium is not available in your country") {
			return ""
		}

		re := regexp.MustCompile(`"INNERTUBE_CONTEXT_GL"\s*:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(body); len(matches) > 1 {
			return strings.ToUpper(matches[1])
		}
	}
	return ""
}

func checkOpenAI(client *http.Client) string {
	req, _ := http.NewRequest("GET", "https://api.openai.com/compliance/cookie_requirements", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		buf := new(strings.Builder)
		io.Copy(buf, resp.Body)
		body := buf.String()

		if !strings.Contains(strings.ToLower(body), "unsupported_country") {
			if checkOpenAIClient(client) {
				region := getOpenAIRegion(client)
				return "GPT⁺-" + region
			}
			return "GPT"
		}
	}
	return ""
}

func checkOpenAIClient(client *http.Client) bool {
	req, _ := http.NewRequest("GET", "https://ios.chat.openai.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6_0 like Mac OS X) AppleWebKit/537.36 Mobile/16G29 ChatGPT/3.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "com.openai.chatgpt")
	req.Header.Set("Origin", "https://chat.openai.com")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	buf := new(strings.Builder)
	io.Copy(buf, resp.Body)
	body := strings.ToLower(buf.String())

	return !strings.Contains(body, "unsupported_country") && !strings.Contains(body, "vpn")
}

func getOpenAIRegion(client *http.Client) string {
	req, _ := http.NewRequest("GET", "https://chat.openai.com/cdn-cgi/trace", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "US"
	}
	defer resp.Body.Close()

	buf := new(strings.Builder)
	io.Copy(buf, resp.Body)

	re := regexp.MustCompile(`loc=([A-Z]{2})`)
	if matches := re.FindStringSubmatch(buf.String()); len(matches) > 1 {
		return matches[1]
	}
	return "US"
}

func checkDisney(client *http.Client) bool {
	req, _ := http.NewRequest("GET", "https://www.disneyplus.com/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func checkGemini(client *http.Client) bool {
	req, _ := http.NewRequest("GET", "https://gemini.google.com/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func writeOutputs(path string, results []UnlockResult, nodes []Node) {
	file, err := os.Create(path)
	if err != nil {
		slog.Error("Failed to create output file", "error", err)
		return
	}
	defer file.Close()

	buf := bufio.NewWriter(file)
	defer buf.Flush()

	header := fmt.Sprintf("# MagicHub Node Check Results\n# Generated: %s\n# Total: %d\n\n",
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

	fmt.Println("\n=== Check Summary ===")
	fmt.Printf("Total nodes: %d\n", stats["total"])
	fmt.Printf("Alive: %d (%.1f%%)\n", stats["alive"], float64(stats["alive"])/float64(stats["total"])*100)
	fmt.Println("\nUnlock Statistics:")
	fmt.Printf("  Netflix: %d\n", stats["netflix"])
	fmt.Printf("  YouTube: %d\n", stats["youtube"])
	fmt.Printf("  OpenAI: %d\n", stats["openai"])
	fmt.Printf("  Disney+: %d\n", stats["disney"])
	fmt.Printf("  Gemini: %d\n", stats["gemini"])
}
