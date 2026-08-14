package rpc_client

import (
	"log"

	shortenerV1 "shorturl/api/shortener/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 全局 gRPC 客户端对象
var ShortenerClient shortenerV1.ShortenerClient

// InitRPCClient 初始化与 transform-rpc 的 gRPC 连接
func InitRPCClient(targetAddr string) {
	conn, err := grpc.NewClient(
		targetAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("连接 transform-rpc 核心服务失败: %v", err)
	}

	// 实例化短链服务的 gRPC 客户端存根！
	ShortenerClient = shortenerV1.NewShortenerClient(conn)
	log.Printf("✅ 网关已成功连接后端 gRPC 核心服务: %s", targetAddr)
}