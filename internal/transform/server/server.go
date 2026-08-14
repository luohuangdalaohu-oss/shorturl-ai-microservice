package server

import (
	"context"

	shortenerV1 "shorturl/api/shortener/v1"
	"shorturl/internal/transform/service"
)

// Server 实现 proto 里的 ShortenerServer 接口
type Server struct {
	shortenerV1.UnimplementedShortenerServer
	svc *service.ShortenerService
}

func NewServer(svc *service.ShortenerService) *Server {
	return &Server{
		svc: svc,
	}
}

// Shorten 实现 gRPC 的长转短接口
func (s *Server) Shorten(ctx context.Context, req *shortenerV1.ShortenRequest) (*shortenerV1.ShortenResponse, error) {
	shortCode, err := s.svc.Shorten(ctx, req.GetOriginalUrl())
	if err != nil {
		return nil, err
	}

	return &shortenerV1.ShortenResponse{
		ShortCode: shortCode,
		ShortUrl:  "http://127.0.0.1:8080/" + shortCode, // 拼接网关域名
	}, nil
}

// Expand 实现 gRPC 的短查长接口
func (s *Server) Expand(ctx context.Context, req *shortenerV1.ExpandRequest) (*shortenerV1.ExpandResponse, error) {
	originalURL, err := s.svc.Expand(ctx, req.GetShortCode())
	if err != nil {
		return nil, err
	}

	return &shortenerV1.ExpandResponse{
		OriginalUrl: originalURL,
	}, nil
}
