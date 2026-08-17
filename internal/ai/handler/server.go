package handler

import (
	"context"

	aiV1 "shorturl/api/ai/v1"
	"shorturl/internal/ai/logic"
)

type Server struct {
	aiV1.UnimplementedAIServiceServer
	logic *logic.SafetyLogic
}

func NewServer(l *logic.SafetyLogic) *Server {
	return &Server{
		logic: l,
	}
}

// CheckURLSafety gRPC 接口实现
func (s *Server) CheckURLSafety(ctx context.Context, req *aiV1.CheckURLSafetyRequest) (*aiV1.CheckURLSafetyResponse, error) {
	return s.logic.CheckURLSafety(ctx, req.GetUrl())
}

// SummarizeURL gRPC 接口实现
func (s *Server) SummarizeURL(ctx context.Context, req *aiV1.SummarizeURLRequest) (*aiV1.SummarizeURLResponse, error) {
	return s.logic.SummarizeURL(ctx, req.GetUrl())
}
