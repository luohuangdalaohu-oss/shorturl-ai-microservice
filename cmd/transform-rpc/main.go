package main

import (
	"fmt"
	"net"

	shortenerV1 "shorturl/api/shortener/v1"
	"shorturl/internal/pkg/snowflake"
	"shorturl/internal/transform/config"
	"shorturl/internal/transform/dao"
	"shorturl/internal/transform/server"
	"shorturl/internal/transform/service"

	"google.golang.org/grpc"
)

func main() {
	// 1. 读取 YAML 配置文件
	cfg, err := config.LoadConfig("configs/transform.yaml")
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 2. 初始化雪花算法发号器
	if err := snowflake.Init(cfg.Snowflake.StartTime, cfg.Snowflake.MachineID); err != nil {
		panic(fmt.Sprintf("初始化雪花算法失败: %v", err))
	}

	// 3. 初始化 MySQL 和 Redis 连接
	d, err := dao.InitDAO(cfg)
	if err != nil {
		panic(fmt.Sprintf("初始化数据库连接失败: %v", err))
	}

	// 4. 组装 Service 业务层和 gRPC Server
	svc := service.NewShortenerService(d)
	srv := server.NewServer(svc)

	// 5. 监听 TCP 端口（如 :8082）
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("监听端口 %s 失败: %v", addr, err))
	}

	// 6. 创建并注册 gRPC 服务
	s := grpc.NewServer()
	shortenerV1.RegisterShortenerServer(s, srv)

	fmt.Printf("🚀 【transform-rpc 短链微服务】启动成功，正在监听 gRPC 端口 %s ...\n", addr)

	// 7. 启动服务
	if err := s.Serve(lis); err != nil {
		panic(fmt.Sprintf("gRPC 服务启动失败: %v", err))
	}
}
