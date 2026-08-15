package main

import (
	"fmt"
	"log"

	"shorturl/internal/gateway/router"
	"shorturl/internal/gateway/rpc_client"

	"github.com/spf13/viper"
)

// 网关配置结构体
type GatewayConfig struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	TransformRPC struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"transform_rpc"`
}

func main() {
	// 1. 读取 gateway.yaml 配置文件
	v := viper.New()
	v.SetConfigFile("configs/gateway.yaml")
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("读取网关配置文件失败: %v", err)
	}

	var cfg GatewayConfig
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("解析网关配置失败: %v", err)
	}

	// 2. 初始化与后端 transform-rpc 核心服务的 gRPC 连接！
	rpc_client.InitRPCClient(cfg.TransformRPC.Addr)

	// 3. 初始化 Gin 路由
	r := router.InitRouter()

	// 4. 启动 HTTP 网关服务（监听 :8080 端口）
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("🌐 【api-gateway 网关服务】启动成功，正在监听 HTTP 端口 %s ...\n", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("网关启动失败: %v", err)
	}
}
