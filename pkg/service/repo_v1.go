package service

import (
	"context"

	pb "github.com/tkeel-io/tkeel/api/repo/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RepoService struct {
	pb.UnimplementedRepoServer
}

func NewRepoService() *RepoService {
	return &RepoService{}
}

func (s *RepoService) CreateRepo(ctx context.Context, req *pb.CreateRepoRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *RepoService) DeleteRepo(ctx context.Context, req *pb.DeleteRepoRequest) (*pb.DeleteRepoResponse, error) {
	return &pb.DeleteRepoResponse{}, nil
}

func (s *RepoService) ListRepo(ctx context.Context, req *emptypb.Empty) (*pb.ListRepoResponse, error) {
	return &pb.ListRepoResponse{}, nil
}

func (s *RepoService) ListRepoInstaller(ctx context.Context, req *pb.ListRepoInstallerRequest) (*pb.ListRepoInstallerResponse, error) {
	return &pb.ListRepoInstallerResponse{}, nil
}

func (s *RepoService) GetRepoInstaller(ctx context.Context, req *pb.GetRepoInstallerRequest) (*pb.GetRepoInstallerResponse, error) {
	return &pb.GetRepoInstallerResponse{}, nil
}
