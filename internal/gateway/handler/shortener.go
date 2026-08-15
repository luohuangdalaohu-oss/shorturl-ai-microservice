package handler

import (
	"context"
	"net/http"
	"time"

	shortenerV1 "shorturl/api/shortener/v1"
	"shorturl/internal/gateway/rpc_client"

	"github.com/gin-gonic/gin"
)

// ShortenReq HTTP 请求入参结构体
type ShortenReq struct {
	URL string `json:"url" binding:"required"`
}

// ShortenHandler 处理 POST /api/v1/shorten 请求
func ShortenHandler(c *gin.Context) {
	var req ShortenReq
	// 1. 解析前端发来的 JSON 数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: url不能为空"})
		return
	}

	// 2. 设置 2 秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 3. 拿起对讲机，通过 gRPC 呼叫后端的 transform-rpc 服务！
	rpcResp, err := rpc_client.ShortenerClient.Shorten(ctx, &shortenerV1.ShortenRequest{
		OriginalUrl: req.URL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成短链失败: " + err.Error()})
		return
	}

	// 4. 将结果以漂亮的 JSON 格式返回给前端！
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"short_code": rpcResp.GetShortCode(),
			"short_url":  rpcResp.GetShortUrl(),
		},
	})
}

// RedirectHandler 处理 GET /:shortCode 浏览器 302 重定向跳转！
func RedirectHandler(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" || shortCode == "favicon.ico" {
		return
	}

	// 1. 设置 1 秒超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 2. 通过 gRPC 问后端：这个短码对应的长链接是啥？
	rpcResp, err := rpc_client.ShortenerClient.Expand(ctx, &shortenerV1.ExpandRequest{
		ShortCode: shortCode,
	})
	if err != nil || rpcResp.GetOriginalUrl() == "" {
		c.String(http.StatusNotFound, "404 - 短链接不存在或已过期")
		return
	}

	// 3. 🔥 核心魔法：向浏览器下达 HTTP 302 临时重定向指令，浏览器瞬间自动跳转！
	c.Redirect(http.StatusFound, rpcResp.GetOriginalUrl())
}
