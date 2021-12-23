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
	repo, err := hub.GetInstance().Get(req.RepoName)
	if err != nil {
		log.Errorf("error hub get repo(%s): %s", req.RepoName, err)
		if errors.Is(err, hub.ErrRepoNotFound) {
			return nil, pb.ErrRepoNotFound()
		}
	}
	installers, err := repo.Search("*")
	if err != nil {
		log.Errorf("error repo(%s) search * get all installer err: %s",
			req.RepoName, err)
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
	repo, err := hub.GetInstance().Get(req.RepoName)
	if err != nil {
		log.Errorf("error hub get repo(%s): %s", req.RepoName, err)
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
		Readme: func() []byte {
			inAn := i.Annotations()
			if inAn == nil {
				return nil
			}
			bIn, ok := inAn[helm.ReadmeFileNameKey]
			if !ok {
				return nil
			}
			b, ok := bIn.([]byte)
			if !ok {
				log.Errorf("error installer(%s) readme invaild type", ib)
				return nil
			}
			return b
		}(),
		ConfigurationSchema: func() []byte {
			inAn := i.Annotations()
			if inAn == nil {
				return nil
			}
			bIn, ok := inAn[helm.ValuesSchemaKey]
			if !ok {
				return nil
			}
			b, ok := bIn.([]byte)
			if !ok {
				log.Errorf("error installer(%s) configuration schema file invaild type", ib)
				return nil
			}
			return b
		}(),
		SchemaType: func() pb.ConfigurationSchemaType {
			inAn := i.Annotations()
			if inAn == nil {
				return pb.ConfigurationSchemaType_JSON
			}
			iIn, ok := inAn["VALUES.SCHEMA.TYPE"]
			if !ok {
				return pb.ConfigurationSchemaType_JSON
			}
			i, ok := iIn.(int)
			if !ok {
				log.Errorf("error installer(%s) configuration schema file invaild type", ib)
				return pb.ConfigurationSchemaType_JSON
			}
			return pb.ConfigurationSchemaType(i)
		}(),
		ConfigurationFile: func() []byte {
			inAn := i.Annotations()
			if inAn == nil {
				return nil
			}
			valueIn, ok := inAn[helm.ValuesKey]
			if !ok {
				return nil
			}
			value, ok := valueIn.([]byte)
			if !ok {
				log.Errorf("error installer(%s) configuration file type invaild", i.Brief().Name)
			}
			return value
		}(),
		Annotations: func() map[string]*anypb.Any {
			inAn := i.Annotations()
			if inAn == nil {
				return nil
			}
			ret := make(map[string]*anypb.Any)
			for k, v := range inAn {
				if k == helm.ReadmeFileNameKey ||
					k == helm.ValuesSchemaKey ||
					k == helm.ValuesKey ||
					k == "VALUES.SCHEMA.TYPE" {
					continue
				}
				b, err := json.Marshal(v)
				if err != nil {
					log.Errorf("error parse installer(%s) annotasions key(%s): %s", i.Brief().Name, k, err)
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
