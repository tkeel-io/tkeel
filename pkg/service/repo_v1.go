package service

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/pkg/errors"

	version "github.com/hashicorp/go-version"
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
		URL:  req.GetUrl().Url,
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
			sort.Sort(repoSort(ret))
			return ret
		}(),
	}, nil
}

func (s *RepoService) ListAllRepoInstaller(ctx context.Context,
	req *pb.ListAllRepoInstallerRequest,
) (*pb.ListAllRepoInstallerResponse, error) {
	repos := hub.GetInstance().List()
	var resList []*repository.InstallerBrief
	for _, v := range repos {
		res, err := v.Search(getReglarStringKeyWords(req.KeyWords))
		if err != nil {
			log.Warnf("get repo(%s) all installer err: %s", v.Info().Name, err)
			continue
		}
		resList = append(resList, res...)
	}
	tmp := make([]*repository.InstallerBrief, 0, len(resList))
	installedNum := 0
	for _, v := range resList {
		if v.State == repository.StateInstalled {
			installedNum++
			if req.Installed {
				tmp = append(tmp, v)
			}
		}
	}
	if req.Installed {
		resList = tmp
	}
	ibList := iBriefList(resList)
	if req.IsDescending {
		sort.Sort(sort.Reverse(ibList))
	} else {
		sort.Sort(ibList)
	}
	total := ibList.Len()
	start, end := getQueryItemsStartAndEnd(int(req.PageNum), int(req.PageSize), total)
	log.Debugf("%d %d", start, end)
	ibList = ibList[start:end]
	return &pb.ListAllRepoInstallerResponse{
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
		InstalledNum: int32(installedNum),
	}, nil
}

func (s *RepoService) ListRepoInstaller(ctx context.Context,
	req *pb.ListRepoInstallerRequest,
) (*pb.ListRepoInstallerResponse, error) {
	repo, err := hub.GetInstance().Get(req.Repo)
	if err != nil {
		log.Errorf("error hub get repo(%s): %s", req.Repo, err)
		if errors.Is(err, hub.ErrRepoNotFound) {
			return nil, pb.ErrRepoNotFound()
		}
	}
	searchWord := getReglarStringKeyWords(req.KeyWords)
	log.Debugf("search words %s -- %s", req.KeyWords, searchWord)
	installers, err := repo.Search(searchWord)
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
	installedNum := 0
	tmp := make(iBriefList, 0, len(ibList))
	for _, v := range ibList {
		if v.State == repository.StateInstalled {
			installedNum++
			if req.Installed {
				tmp = append(tmp, v)
			}
		}
	}
	if req.Installed {
		ibList = tmp
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
		InstalledNum: int32(installedNum),
	}, nil
}

func (s *RepoService) GetRepoInstaller(ctx context.Context,
	req *pb.GetRepoInstallerRequest,
) (*pb.GetRepoInstallerResponse, error) {
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
	if r == nil {
		return &pb.RepoObject{}
	}
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
		InstallerNum: int32(r.Len()),
	}
}

func convertInstallerBrief2PB(ib *repository.InstallerBrief) *pb.InstallerObject {
	return &pb.InstallerObject{
		Name:    ib.Name,
		Version: ib.Version,
		Repo:    ib.Repo,
		State: func() pb.InstallerState {
			switch ib.State {
			case repository.StateUninstall:
				return pb.InstallerState_UNINSTALL
			case repository.StateInstalled:
				return pb.InstallerState_INSTALLED
			case repository.StateSameNameInstalled:
				return pb.InstallerState_SAME_NAME
			}
			return pb.InstallerState_UNINSTALL
		}(),
		Annotations: ib.Annotations,
		Maintainers: func() []*pb.InstallerObjectMaintainer {
			ret := make([]*pb.InstallerObjectMaintainer, 0, len(ib.Maintainers))
			for _, v := range ib.Maintainers {
				ret = append(ret, &pb.InstallerObjectMaintainer{
					Name:  v.Name,
					Email: v.Email,
					Url:   v.URL,
				})
			}
			return ret
		}(),
		Desc:      ib.Desc,
		Timestamp: uint64(ib.CreateTimestamp),
		Icon:      ib.Icon,
	}
}

func convertInstaller2PB(i repository.Installer) *pb.InstallerObject {
	ib := i.Brief()
	if ib == nil {
		return nil
	}
	return &pb.InstallerObject{
		Name:    ib.Name,
		Version: ib.Version,
		Repo:    ib.Repo,
		State: func() pb.InstallerState {
			switch ib.State {
			case repository.StateUninstall:
				return pb.InstallerState_UNINSTALL
			case repository.StateInstalled:
				return pb.InstallerState_INSTALLED
			case repository.StateSameNameInstalled:
				return pb.InstallerState_SAME_NAME
			}
			return pb.InstallerState_UNINSTALL
		}(),
		Metadata: pbMetadata(i),
		Annotations: func() map[string]string {
			ret := make(map[string]string)
			for k, v := range i.Annotations() {
				if k != repository.ConfigurationKey &&
					k != repository.ConfigurationSchemaKey &&
					k != helm.ReadmeKey &&
					k != helm.ChartDescKey &&
					k != helm.ChartMetaDataKey {
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
		Maintainers: func() []*pb.InstallerObjectMaintainer {
			ret := make([]*pb.InstallerObjectMaintainer, 0, len(i.Brief().Maintainers))
			for _, v := range i.Brief().Maintainers {
				ret = append(ret, &pb.InstallerObjectMaintainer{
					Name:  v.Name,
					Email: v.Email,
					Url:   v.URL,
				})
			}
			return ret
		}(),
		Desc:      i.Brief().Desc,
		Timestamp: uint64(ib.CreateTimestamp),
		Icon:      ib.Icon,
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
	if ib[i].Name != ib[j].Name {
		return ib[i].Name < ib[j].Name
	}
	iVer, err := version.NewVersion(ib[i].Version)
	if err != nil {
		return true
	}
	jVer, err := version.NewVersion(ib[j].Version)
	if err != nil {
		return false
	}
	if !iVer.Equal(jVer) {
		return iVer.LessThan(jVer)
	}
	if ib[i].CreateTimestamp != ib[j].CreateTimestamp {
		return ib[i].CreateTimestamp < ib[j].CreateTimestamp
	}
	return ib[i].Repo < ib[j].Repo
}

func (ib iBriefList) Swap(i, j int) {
	ib[i], ib[j] = ib[j], ib[i]
}

type repoSort []*pb.RepoObject

func (a repoSort) Len() int           { return len(a) }
func (a repoSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a repoSort) Less(i, j int) bool { return a[i].Name < a[j].Name }
