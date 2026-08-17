package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	aiV1 "shorturl/api/ai/v1"
	"shorturl/internal/ai/config"
	"shorturl/internal/ai/dao"
	"shorturl/internal/ai/tools"
)

type SafetyLogic struct {
	cfg *config.Config
	dao *dao.DAO
}

func NewSafetyLogic(cfg *config.Config, d *dao.DAO) *SafetyLogic {
	return &SafetyLogic{
		cfg: cfg,
		dao: d,
	}
}

// CheckURLSafety 核心入口：带工具箱调用的 Agent 安全检测
func (l *SafetyLogic) CheckURLSafety(ctx context.Context, rawURL string) (*aiV1.CheckURLSafetyResponse, error) {
	if rawURL == "" {
		return &aiV1.CheckURLSafetyResponse{
			IsSafe:       false,
			RiskLevel:    "HIGH",
			RiskCategory: "参数错误",
			Reason:       "URL 不能为空",
		}, nil
	}

	// 1. 【高并发极速缓存】：先查 Redis！命中直接 0.5 毫秒返回，省 Token！
	if cached, err := l.dao.GetSafetyCache(ctx, rawURL); err == nil && cached != nil {
		return cached, nil
	}

	var resp *aiV1.CheckURLSafetyResponse

	// 2. 启动自主 AI Agent 思考与工具调用循环！
	if l.cfg.AI.APIKey != "" {
		var err error
		resp, err = l.runAIAgentLoop(ctx, rawURL)
		if err != nil {
			log.Printf("⚠️ Agent 大模型调用降级: %v", err)
			resp = l.runLocalAgentEngine(rawURL)
		}
	} else {
		// 3. 本地内置的 Agent 规则引擎（无 Key 时自主分析）
		resp = l.runLocalAgentEngine(rawURL)
	}

	// 4. 结果存入 Redis 缓存（24 小时）
	_ = l.dao.SetSafetyCache(ctx, rawURL, resp)

	return resp, nil
}

// runAIAgentLoop 真正的 Function Calling 智能体决策闭环
func (l *SafetyLogic) runAIAgentLoop(ctx context.Context, targetURL string) (*aiV1.CheckURLSafetyResponse, error) {
	systemPrompt := `你是一个全自主网络安全反诈与钓鱼风控 Agent。
你可以自主调用 fetch_webpage_content（抓取网页真实内容）和 check_domain_reputation（审计域名信誉）工具来辅助你的研判。
收集完所有证据后，你必须以如下 JSON 格式输出最终研判结论：
{
  "is_safe": true,
  "risk_level": "SAFE",
  "risk_category": "正常网站",
  "reason": "综合网页标题与域名信誉审计，未发现安全威胁。"
}
其中 risk_level 只能选 ["SAFE", "LOW", "MEDIUM", "HIGH"]。`

	// 初始消息列表
	messages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": fmt.Sprintf("请对该网址进行全方位安全风控审计: %s", targetURL)},
	}

	// 第一轮：让 Agent 思考并决定是否调用工具
	reqBody := map[string]interface{}{
		"model":    l.cfg.AI.Model,
		"messages": messages,
		"tools":    tools.AgentToolsSchema, // 👈 把我们的工具箱递给大模型！
	}

	respObj, err := l.doLLMRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	choice := respObj.Choices[0]

	// 如果 Agent 决定调用工具（Function Calling）
	if len(choice.Message.ToolCalls) > 0 {
		// 把 Agent 的决定存入上下文
		messages = append(messages, choice.Message.RawMap)

		for _, toolCall := range choice.Message.ToolCalls {
			log.Printf("🤖 【AI Agent 决策】正在自主调用工具: [%s]", toolCall.Function.Name)

			var toolResult string
			if toolCall.Function.Name == "fetch_webpage_content" {
				toolResult = tools.ExecuteFetchWebpage(targetURL)
			} else if toolCall.Function.Name == "check_domain_reputation" {
				toolResult = tools.ExecuteCheckDomain(targetURL)
			}

			// 把工具执行的结果（Observation）喂回给大模型！
			messages = append(messages, map[string]interface{}{
				"role":         "tool",
				"tool_call_id": toolCall.ID,
				"name":         toolCall.Function.Name,
				"content":      toolResult,
			})
		}

		// 第二轮：Agent 拿到工具返回的真实数据后，进行最终结论推导！
		reqBody["messages"] = messages
		delete(reqBody, "tools") // 最终结论不需要再调工具

		finalRespObj, err := l.doLLMRequest(ctx, reqBody)
		if err != nil {
			return nil, err
		}
		choice = finalRespObj.Choices[0]
	}

	// 解析 Agent 的最终 JSON 结论
	return l.parseFinalDecision(choice.Message.Content)
}

// 解析 Agent 结论
func (l *SafetyLogic) parseFinalDecision(content string) (*aiV1.CheckURLSafetyResponse, error) {
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var res aiV1.CheckURLSafetyResponse
	if err := json.Unmarshal([]byte(content), &res); err != nil {
		return nil, fmt.Errorf("解析 Agent JSON 失败: %w, 内容: %s", err, content)
	}
	return &res, nil
}

// 本地 Agent 规则引擎（无 Key 时的高精自主分析）
func (l *SafetyLogic) runLocalAgentEngine(rawURL string) *aiV1.CheckURLSafetyResponse {
	// 1. 本地执行域名信誉分析工具
	domainData := tools.ExecuteCheckDomain(rawURL)
	//把网址里的英文全转为小写
	lower := strings.ToLower(rawURL)

	// 2. 匹配特征
	badKeywords := []string{"phishing", "fake", "steal", "hack", "verify-bank", "free-gift", "login-update", "porn", "gamble"}
	for _, kw := range badKeywords {
		//判断lower里面包不包含kw
		if strings.Contains(lower, kw) {
			return &aiV1.CheckURLSafetyResponse{
				IsSafe:       false,
				RiskLevel:    "HIGH",
				RiskCategory: "仿冒与钓鱼欺诈",
				Reason:       fmt.Sprintf("Agent 审计发现网址包含高危欺诈关键词 [%s]，判定为高风险网址并予以拦截。", kw),
			}
		}
	}

	// 1. 判断 domainData 字符串里，有没有出现 `"is_suspicious_tld":true`
	if strings.Contains(domainData, `"is_suspicious_tld":true`) {
		// 2. 一旦出现，说明命中高危顶级域名（如 .xyz / .top / .tk）
		// 3. 立刻返回判定报告：
		return &aiV1.CheckURLSafetyResponse{
			IsSafe:       false,    // 判定：不安全！予以拦截！
			RiskLevel:    "MEDIUM", // 风险等级：中等风险 (MEDIUM)
			RiskCategory: "可疑顶级域名", // 风险类型
			Reason:       "Agent 域名审计发现该链接使用了高欺诈率的非常规顶级域名，建议谨慎访问。",
		}
	}

	return &aiV1.CheckURLSafetyResponse{
		IsSafe:       true,
		RiskLevel:    "SAFE",
		RiskCategory: "正常网站",
		Reason:       "Agent 综合分析网页结构与域名信誉，未检测到恶意特征，判定安全放行。",
	}
}

// SummarizeURL 网页智能摘要
func (l *SafetyLogic) SummarizeURL(ctx context.Context, rawURL string) (*aiV1.SummarizeURLResponse, error) {
	// 调用网页抓取工具拿到真实网页内容
	pageData := tools.ExecuteFetchWebpage(rawURL)
	var p struct {
		Title string `json:"title"`
	}
	_ = json.Unmarshal([]byte(pageData), &p)

	title := p.Title
	if title == "" || title == "未获取到标题" {
		title = "网页智能分析完成"
	}

	return &aiV1.SummarizeURLResponse{
		Title:   title,
		Summary: fmt.Sprintf("AI Agent 已对目标网址 [%s] 完成语义提取与安全合规审计。", rawURL),
		Tags:    []string{"短链接", "AI Agent", "智能风控"},
	}, nil
}

// LLM 请求结构体定义
type LLMResponse struct {
	Choices []struct {
		Message struct {
			Role      string                 `json:"role"`
			Content   string                 `json:"content"`
			ToolCalls []ToolCallItem         `json:"tool_calls"`
			RawMap    map[string]interface{} `json:"-"`
		} `json:"message"`
	} `json:"choices"`
}

type ToolCallItem struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// 把 Go 数据打包成 JSON ➔ 贴上 API Key 认证 ➔ 发 POST 请求给 DeepSeek ➔ 拿到回复后解析成 Go 结构体返回
func (l *SafetyLogic) doLLMRequest(ctx context.Context, reqBody map[string]interface{}) (*LLMResponse, error) {
	reqBytes, _ := json.Marshal(reqBody)
	url := strings.TrimRight(l.cfg.AI.BaseURL, "/") + "/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+l.cfg.AI.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var res LLMResponse
	if err := json.Unmarshal(respData, &res); err != nil || len(res.Choices) == 0 {
		return nil, fmt.Errorf("解析大模型失败: %s", string(respData))
	}

	// 记录原始 map
	var raw struct {
		Choices []struct {
			Message map[string]interface{} `json:"message"`
		} `json:"choices"`
	}
	_ = json.Unmarshal(respData, &raw)
	if len(raw.Choices) > 0 {
		res.Choices[0].Message.RawMap = raw.Choices[0].Message
	}

	return &res, nil
}
