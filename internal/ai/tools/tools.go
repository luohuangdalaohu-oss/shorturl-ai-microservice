package tools

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// 1. 定义暴露给大模型的 Tool 规范描述（OpenAI/DeepSeek 标准格式）
var AgentToolsSchema = []map[string]interface{}{
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "fetch_webpage_content",
			"description": "实时抓取目标网页的真实 Title 标题、Meta 描述和前文正文，用于识别网页真实意图和诈骗仿冒内容",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "要抓取的完整网页 URL 地址",
					},
				},
				"required": []string{"url"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "check_domain_reputation",
			"description": "深度分析域名的结构特征（如是否为纯IP地址、可疑顶级域名、伪装子域名等）",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain": map[string]interface{}{
						"type":        "string",
						"description": "待分析的域名（如 bad-bank.login.xyz）",
					},
				},
				"required": []string{"domain"},
			},
		},
	},
}

// 2. 工具 ① 的真实执行逻辑：网页抓取
// (用爬虫技术去目标网址读 5KB 数据，把网页的真实标题 <title> 抓出来，打包成 JSON 字符串交付给大模型去审计)
func ExecuteFetchWebpage(targetURL string) string {
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 防止有些网站证书过期
		},
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return fmt.Sprintf(`{"error": "无法创建请求: %v"}`, err)
	}

	// 伪装请求头（User-Agent）：把自己伪装成 Windows 电脑上的 Chrome 浏览器，防止被目标网站当成爬虫直接拦截
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf(`{"status": "fail", "reason": "网页无法访问或连接超时: %v"}`, err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 5000)) // 只读前 5KB 防止大文件爆内存
	bodyStr := string(bodyBytes)

	// 正则提取 <title>
	title := "未获取到标题"
	reTitle := regexp.MustCompile(`(?i)<title>(.*?)</title>`)
	if match := reTitle.FindStringSubmatch(bodyStr); len(match) > 1 {
		title = strings.TrimSpace(match[1])
	}

	res := map[string]interface{}{
		"status_code":  resp.StatusCode,
		"title":        title,
		"body_snippet": strings.Join(strings.Fields(bodyStr)[:min(len(strings.Fields(bodyStr)), 50)], " "),
	}
	out, _ := json.Marshal(res)
	return string(out)
}

// 工具 ② 的真实执行逻辑：域名信誉审计
// (从网址中提取域名，检测它是不是高危后缀（如 .xyz）、是不是多级嵌套伪装域名、是不是纯数字 IP 裸奔，并将特征数据打包成 JSON 返回)
func ExecuteCheckDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	domain := rawURL
	if err == nil && u.Host != "" {
		domain = u.Host
	}

	isSuspiciousTLD := false
	suspiciousTLDs := []string{".top", ".xyz", ".club", ".vip", ".cc", ".tk", ".work"}
	for _, tld := range suspiciousTLDs {
		if strings.HasSuffix(strings.ToLower(domain), tld) {
			isSuspiciousTLD = true
			break
		}
	}

	subdomainCount := strings.Count(domain, ".")
	res := map[string]interface{}{
		"domain":            domain,
		"is_suspicious_tld": isSuspiciousTLD,
		"subdomain_depth":   subdomainCount,
		"is_raw_ip":         regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+`).MatchString(domain),
	}
	out, _ := json.Marshal(res)
	return string(out)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
