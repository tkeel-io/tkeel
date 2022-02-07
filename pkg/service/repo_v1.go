package service

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel/api/repo/v1"
	"github.com/tkeel-io/tkeel/pkg/hub"
	"github.com/tkeel-io/tkeel/pkg/repository"
	"github.com/tkeel-io/tkeel/pkg/repository/helm"
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
	ibList := iBriefList(installers)
	if req.IsDescending {
		sort.Sort(sort.Reverse(ibList))
	} else {
		sort.Sort(ibList)
	}
	if req.Installed {
		for i, v := range ibList {
			if !v.Installed {
				ibList = ibList[:i]
				break
			}
		}
	}
	total := ibList.Len()
	start, end := getQueryItemsStartAndEnd(int(req.PageNum), int(req.PageSize), total)
	log.Debugf("%d %d", start, end)
	ibList = ibList[start:end]
	return &pb.ListRepoInstallerResponse{
		Total:    int32(total),
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		BriefInstallers: func() []*pb.InstallerObject {
			ret := make([]*pb.InstallerObject, 0, len(ibList))
			for _, v := range ibList {
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
		if errors.Is(err, helm.ErrNotFound) {
			return nil, pb.ErrInstallerNotFound()
		}
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
		Metadata: func() map[string][]byte {
			ret := make(map[string][]byte)
			return ret
		}(),
		Annotations: func() map[string]string {
			ret := make(map[string]string)
			for k, v := range r.Info().Annotations {
				b, err := json.Marshal(v)
				if err != nil {
					log.Errorf("error parse installer(%s) annotasions key(%s): %s", r.Info().Name, k, err)
					continue
				}
				ret[k] = string(b)
			}
			return ret
		}(),
	}
}

func convertInstallerBrief2PB(ib *repository.InstallerBrief) *pb.InstallerObject {
	return &pb.InstallerObject{
		Name:        ib.Name,
		Version:     ib.Version,
		Repo:        ib.Repo,
		Installed:   ib.Installed,
		Annotations: ib.Annotations,
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
		Annotations: func() map[string]string {
			ret := make(map[string]string)
			for k, v := range i.Annotations() {
				if k != repository.ConfigurationKey &&
					k != repository.ConfigurationSchemaKey &&
					k != helm.ReadmeKey &&
					k != helm.ChartDescKey {
					if vstr, ok := v.(string); ok {
						ret[k] = vstr
						continue
					}
					b, err := json.Marshal(v)
					if err != nil {
						log.Errorf("error parse installer(%s) annotasions key(%s): %s", i.Brief().Name, k, err)
						continue
					}
					ret[k] = string(b)
				}
			}
			return ret
		}(),
	}
}

func pbMetadata(i repository.Installer) map[string][]byte {
	anno := i.Annotations()
	ret := make(map[string][]byte, len(anno))
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
			ret[k] = vb
		}
	}
	return ret
}

type iBriefList []*repository.InstallerBrief

func (ib iBriefList) Len() int {
	return len(ib)
}

func (ib iBriefList) Less(i, j int) bool {
	if ib[i].Installed != ib[j].Installed {
		return ib[j].Installed
	}
	if ib[i].Name != ib[j].Name {
		return ib[i].Name < ib[j].Name
	}
	iVer, err := strconv.ParseFloat(ib[i].Version, 64)
	if err != nil {
		return true
	}
	jVer, err := strconv.ParseFloat(ib[j].Version, 64)
	if err != nil {
		return false
	}
	return iVer < jVer
}

func (ib iBriefList) Swap(i, j int) {
	ib[i], ib[j] = ib[j], ib[i]
}
