package handler

import (
	"context"
	"net/http"
	"time"

	aiV1 "shorturl/api/ai/v1"
	shortenerV1 "shorturl/api/shortener/v1"
	"shorturl/internal/gateway/rpc_client"

	"github.com/gin-gonic/gin"
)

type ShortenReq struct {
	URL string `json:"url" binding:"required"`
}

// ShortenHandler 处理生成短链请求（带 AI 智能风控反诈）
func ShortenHandler(c *gin.Context) {
	var req ShortenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: url不能为空"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	// 🛡️ 🔥 第一步：呼叫 AI Agent 进行全方位反诈安全与钓鱼风控审计！
	safetyResp, err := rpc_client.AIClient.CheckURLSafety(ctx, &aiV1.CheckURLSafetyRequest{
		Url: req.URL,
	})
	if err == nil && safetyResp != nil && !safetyResp.GetIsSafe() {
		// 🚫 如果 AI Agent 判定为高危恶意/钓鱼网址，直接予以熔断拦截！
		c.JSON(http.StatusForbidden, gin.H{
			"code": 403,
			"msg":  "【AI 安全智能体拦截】" + safetyResp.GetReason(),
			"data": gin.H{
				"risk_level":    safetyResp.GetRiskLevel(),
				"risk_category": safetyResp.GetRiskCategory(),
				"reason":        safetyResp.GetReason(),
			},
		})
		return
	}

	// ⚡ 第二步：AI 审计安全通过，调用短链核心服务生成短链！
	rpcResp, err := rpc_client.ShortenerClient.Shorten(ctx, &shortenerV1.ShortenRequest{
		OriginalUrl: req.URL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成短链失败: " + err.Error()})
		return
	}

	// 第三步：返回生成的短链及 AI 审核报告！
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"short_code": rpcResp.GetShortCode(),
			"short_url":  rpcResp.GetShortUrl(),
			"ai_security": gin.H{
				"is_safe":       true,
				"risk_level":    safetyResp.GetRiskLevel(),
				"audit_summary": safetyResp.GetReason(),
			},
		},
	})
}

// RedirectHandler 浏览器 302 重定向
func RedirectHandler(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" || shortCode == "favicon.ico" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rpcResp, err := rpc_client.ShortenerClient.Expand(ctx, &shortenerV1.ExpandRequest{
		ShortCode: shortCode,
	})
	if err != nil || rpcResp.GetOriginalUrl() == "" {
		c.String(http.StatusNotFound, "404 - 短链接不存在或已过期")
		return
	}

	// HTTP 302 临时重定向
	c.Redirect(http.StatusFound, rpcResp.GetOriginalUrl())
}
