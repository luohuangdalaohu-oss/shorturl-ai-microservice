package rpc_client

import (
	"log"

	aiV1 "shorturl/api/ai/v1"
	shortenerV1 "shorturl/api/shortener/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 全局 gRPC 客户端存根对象
var (
	ShortenerClient shortenerV1.ShortenerClient
	AIClient        aiV1.AIServiceClient // 👈 AI 微服务专属小秘书！
)

// InitRPCClient 初始化网关对所有下游微服务的连接
func InitRPCClient(transformAddr, aiAddr string) {
	// 1. 连接 transform-rpc 核心服务
	connTransform, err := grpc.NewClient(
		transformAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("连接 transform-rpc 核心服务失败: %v", err)
	}
	ShortenerClient = shortenerV1.NewShortenerClient(connTransform)
	log.Printf("✅ 网关已成功连接后端短链核心服务: %s", transformAddr)

	// 2. 连接 ai-service AI 智能体服务
	connAI, err := grpc.NewClient(
		aiAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("连接 ai-service 智能服务失败: %v", err)
	}
	AIClient = aiV1.NewAIServiceClient(connAI)
	log.Printf("🤖 网关已成功连接后端 AI 智能微服务: %s", aiAddr)
}
