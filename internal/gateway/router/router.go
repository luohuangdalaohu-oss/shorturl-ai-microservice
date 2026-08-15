package router

import (
	"net/http"

	"shorturl/internal/gateway/handler"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化 Gin 路由引擎
func InitRouter() *gin.Engine {
	r := gin.Default()

	// 1. 简单的健康检查接口
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "service": "api-gateway"})
	})

	// 2. 短链生成 API 路由组
	v1 := r.Group("/api/v1")
	{
		v1.POST("/shorten", handler.ShortenHandler) // 生成短链
	}

	// 3. 根路径短码重定向（最关键的 302 跳转路由！）
	r.GET("/:shortCode", handler.RedirectHandler)

	return r
}
