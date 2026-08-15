package handler

import (
	"context"

	shortenerV1 "shorturl/api/shortener/v1"
	"shorturl/internal/transform/logic"
)

// Server 实现 gRPC 的 ShortenerServer 接口（Handler 协议处理层）
type Server struct {
	shortenerV1.UnimplementedShortenerServer
	logic *logic.ShortenerLogic
}

func NewServer(l *logic.ShortenerLogic) *Server {
	return &Server{
		logic: l,
	}
}

// Shorten 处理长转短 gRPC 请求
func (s *Server) Shorten(ctx context.Context, req *shortenerV1.ShortenRequest) (*shortenerV1.ShortenResponse, error) {
	shortCode, err := s.logic.Shorten(ctx, req.GetOriginalUrl())
	if err != nil {
		return nil, err
	}

	return &shortenerV1.ShortenResponse{
		ShortCode: shortCode,
		ShortUrl:  "http://127.0.0.1:8080/" + shortCode,
	}, nil
}

// Expand 处理短查长 gRPC 请求
func (s *Server) Expand(ctx context.Context, req *shortenerV1.ExpandRequest) (*shortenerV1.ExpandResponse, error) {
	originalURL, err := s.logic.Expand(ctx, req.GetShortCode())
	if err != nil {
		return nil, err
	}

	return &shortenerV1.ExpandResponse{
		OriginalUrl: originalURL,
	}, nil
}
