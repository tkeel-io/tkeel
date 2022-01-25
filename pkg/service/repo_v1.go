package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel/api/repo/v1"
	"github.com/tkeel-io/tkeel/pkg/hub"
	"github.com/tkeel-io/tkeel/pkg/repository"
	"github.com/tkeel-io/tkeel/pkg/repository/helm"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RepoService struct {
	pb.UnimplementedRepoServer
}

func NewRepoService() *RepoService {
	return &RepoService{}
}

func (s *RepoService) CreateRepo(ctx context.Context, req *pb.CreateRepoRequest) (*emptypb.Empty, error) {
	info := &repository.Info{
		Name: req.Name,
		URL:  req.Url,
		// TODO: add annotations.
	}
	if err := hub.GetInstance().Add(info); err != nil {
		log.Errorf("error hub add repo(%s): %s", info, err)
		if errors.Is(err, hub.ErrRepoExist) {
			return nil, pb.ErrRepoExist()
		}
		return nil, pb.ErrInternalError()
	}
	return &emptypb.Empty{}, nil
}

func (s *RepoService) DeleteRepo(ctx context.Context, req *pb.DeleteRepoRequest) (*pb.DeleteRepoResponse, error) {
	repo, err := hub.GetInstance().Delete(req.Name)
	if err != nil {
		log.Errorf("error hub delete repo(%s): %s", req.Name, err)
		if errors.Is(err, hub.ErrRepoNotFound) {
			return nil, pb.ErrRepoNotFound()
		}
	}
	return &pb.DeleteRepoResponse{
		Repo: convertRepo2PB(repo),
	}, nil
}

func (s *RepoService) ListRepo(ctx context.Context, req *emptypb.Empty) (*pb.ListRepoResponse, error) {
	repoList := hub.GetInstance().List()
	return &pb.ListRepoResponse{
		Repos: func() []*pb.RepoObject {
			ret := make([]*pb.RepoObject, 0, len(repoList))
			for _, v := range repoList {
				ret = append(ret, convertRepo2PB(v))
			}
			return ret
		}(),
	}, nil
}

func (s *RepoService) ListRepoInstaller(ctx context.Context,
	req *pb.ListRepoInstallerRequest) (*pb.ListRepoInstallerResponse, error) {
	repo, err := hub.GetInstance().Get(req.Repo)
	if err != nil {
		log.Errorf("error hub get repo(%s): %s", req.Repo, err)
		if errors.Is(err, hub.ErrRepoNotFound) {
			return nil, pb.ErrRepoNotFound()
		}
	}
	installers, err := repo.Search("*")
	if err != nil {
		log.Errorf("error repo(%s) search * get all installer err: %s",
			req.Repo, err)
		return nil, pb.ErrInternalError()
	}
	return &pb.ListRepoInstallerResponse{
		BriefInstallers: func() []*pb.InstallerObject {
			ret := make([]*pb.InstallerObject, 0, len(installers))
			for _, v := range installers {
				ret = append(ret, convertInstallerBrief2PB(v))
			}
			return ret
		}(),
	}, nil
}

func (s *RepoService) GetRepoInstaller(ctx context.Context,
	req *pb.GetRepoInstallerRequest) (*pb.GetRepoInstallerResponse, error) {
	repo, err := hub.GetInstance().Get(req.Repo)
	if err != nil {
		log.Errorf("error hub get repo(%s): %s", req.Repo, err)
		if errors.Is(err, hub.ErrRepoNotFound) {
			return nil, pb.ErrRepoNotFound()
		}
	}
	installer, err := repo.Get(req.InstallerName, req.InstallerVersion)
	if err != nil {
		log.Errorf("error repo(%s) get installer(%s/%s): %s",
			repo.Info(), req.InstallerName, req.InstallerVersion, err)
		return nil, pb.ErrInvalidArgument()
	}
	return &pb.GetRepoInstallerResponse{
		Installer: convertInstaller2PB(installer),
	}, nil
}

func convertRepo2PB(r repository.Repository) *pb.RepoObject {
	return &pb.RepoObject{
		Name: r.Info().Name,
		Url:  r.Info().URL,
		Annotations: func() map[string]*anypb.Any {
			ret := make(map[string]*anypb.Any)
			for k, v := range r.Info().Annotations {
				b, err := json.Marshal(v)
				if err != nil {
					log.Errorf("error parse installer(%s) annotasions key(%s): %s", r.Info().Name, k, err)
					continue
				}
				ret[k] = &anypb.Any{
					Value: b,
				}
			}
			return ret
		}(),
	}
}

func convertInstallerBrief2PB(ib *repository.InstallerBrief) *pb.InstallerObject {
	return &pb.InstallerObject{
		Name:      ib.Name,
		Version:   ib.Version,
		Repo:      ib.Repo,
		Installed: ib.Installed,
	}
}

func convertInstaller2PB(i repository.Installer) *pb.InstallerObject {
	ib := i.Brief()
	if ib == nil {
		return nil
	}
	return &pb.InstallerObject{
		Name:      ib.Name,
		Version:   ib.Version,
		Repo:      ib.Repo,
		Installed: ib.Installed,
		Metadata:  pbMetadata(i),
	}
}

func pbMetadata(i repository.Installer) map[string]*anypb.Any {
	anno := i.Annotations()
	ret := make(map[string]*anypb.Any, len(anno))
	for k, v := range anno {
		if k == repository.ConfigurationKey ||
			k == repository.ConfigurationSchemaKey ||
			k == helm.ReadmeKey ||
			k == helm.ChartDescKey {
			vb, ok := v.([]byte)
			if !ok {
				log.Errorf("installer(%s) annotasion(%s) is invalid type", i.Brief(), k)
				continue
			}
			ret[k] = &anypb.Any{
				Value: vb,
			}
		}
	}
	return ret
}
