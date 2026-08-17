package main

import (
	"fmt"
	"log"

	"shorturl/internal/gateway/router"
	"shorturl/internal/gateway/rpc_client"

	"github.com/spf13/viper"
)

type GatewayConfig struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	TransformRPC struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"transform_rpc"`
	AIRPC struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"ai_rpc"`
}

func main() {
	// 1. 读取 gateway.yaml 配置
	v := viper.New()
	v.SetConfigFile("configs/gateway.yaml")
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("读取网关配置文件失败: %v", err)
	}

	var cfg GatewayConfig
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("解析网关配置失败: %v", err)
	}

	// 2. 同时连接后端短链核心服务 (8082) 和 AI 智能体服务 (8084)！
	rpc_client.InitRPCClient(cfg.TransformRPC.Addr, cfg.AIRPC.Addr)

	// 3. 启动 Gin 路由
	r := router.InitRouter()

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("🌐 【api-gateway 网关服务】启动成功，正在监听 HTTP 端口 %s ...\n", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("网关启动失败: %v", err)
	}
}
