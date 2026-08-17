package main

import (
	"fmt"
	"net"

	aiV1 "shorturl/api/ai/v1"
	"shorturl/internal/ai/config"
	"shorturl/internal/ai/dao"
	"shorturl/internal/ai/handler"
	"shorturl/internal/ai/logic"

	"google.golang.org/grpc"
)

func main() {
	// 1. 读取 YAML 配置
	cfg, err := config.LoadConfig("configs/ai.yaml")
	if err != nil {
		panic(fmt.Sprintf("加载 AI 配置失败: %v", err))
	}

	// 2. 初始化 Redis DAO 缓存层
	d, err := dao.InitDAO(cfg)
	if err != nil {
		panic(fmt.Sprintf("初始化 AI Redis 缓存失败: %v", err))
	}

	// 3. 组装 Logic 业务逻辑层与 Handler 协议层
	l := logic.NewSafetyLogic(cfg, d)
	h := handler.NewServer(l)

	// 4. 监听 TCP 端口（:8084）
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("监听 AI 端口 %s 失败: %v", addr, err))
	}

	// 5. 注册并启动 gRPC 服务！
	s := grpc.NewServer()
	aiV1.RegisterAIServiceServer(s, h)

	fmt.Printf("🤖 【ai-service AI 智能体微服务】启动成功，正在监听 gRPC 端口 %s ...\n", addr)

	if err := s.Serve(lis); err != nil {
		panic(fmt.Sprintf("AI gRPC 服务启动失败: %v", err))
	}
}
