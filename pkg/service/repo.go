package service

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/tkeel/api/repo/v1"
	"github.com/tkeel-io/tkeel/pkg/helm"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RepoService struct {
	pb.UnimplementedRepoServer
}

func NewRepoService() *RepoService {
	return &RepoService{}
}

func (s *RepoService) CreateRepo(ctx context.Context, req *pb.CreateRepoRequest) (*emptypb.Empty, error) {
	if req.Addr == "" {
		return nil, pb.ErrInvalidArgument()
	}
	if err := helm.AddRepo(req.Addr); err != nil {
		err = errors.Wrap(err, "add helm repo err")
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *RepoService) DeleteRepo(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if err := helm.DeleteOwnRepo(); err != nil {
		err = errors.Wrap(err, "call repo remove failed")
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *RepoService) ListRepo(ctx context.Context, req *emptypb.Empty) (*pb.ListRepoResponse, error) {
	c, err := helm.ListRepo("json")
	if err != nil {
		err = errors.Wrap(err, "get repo list err")
		return nil, err
	}
	return &pb.ListRepoResponse{List: string(c)}, nil
}

func (s *RepoService) InstallPluginFromRepo(ctx context.Context, req *pb.InstallPluginFromRepoRequest) (*emptypb.Empty, error) {
	chart := strings.Join([]string{req.Repo, req.Plugin}, "/")
	if err := helm.Install(ctx, req.Name, chart, req.Version); err != nil {
		switch {
		case errors.Is(err, helm.ErrVersionPattern):
			return nil, pb.ErrInvalidArgument()
		}
		err = errors.Wrap(err, "install plugin err")
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
