package service

import (
	"context"

	pb "github.com/tkeel-io/tkeel/api/security_oauth/v1"
)

type OauthService struct {
	pb.UnimplementedOauthServer
}

func NewOauthService() *OauthService {
	return &OauthService{}
}

func (s *OauthService) Authorize(ctx context.Context, req *pb.AuthorizeRequest) (*pb.AuthorizeResponse, error) {
	return &pb.AuthorizeResponse{}, nil
}
func (s *OauthService) Token(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	return &pb.TokenResponse{}, nil
}
