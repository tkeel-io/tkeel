package util

import (
	pb "github.com/tkeel-io/tkeel/api/plugin/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/repository"
)

func ConvertModel2PluginObjectPb(p *model.Plugin, pr *model.PluginRoute) *pb.PluginObject {
	return &pb.PluginObject{
		Id:                p.ID,
		PluginVersion:     p.PluginVersion,
		TkeelVersion:      p.TkeelVersion,
		AddonsPoint:       p.AddonsPoint,
		ImplementedPlugin: p.ImplementedPlugin,
		Secret: &pb.Secret{
			Data: p.Secret.Data,
		},
		RegisterTimestamp: p.RegisterTimestamp,
		ActiveTenantes:    p.ActiveTenantes,
		RegisterAddons: func() []*pb.RegisterAddons {
			ret := make([]*pb.RegisterAddons, 0, len(pr.RegisterAddons))
			for k, v := range pr.RegisterAddons {
				ret = append(ret, &pb.RegisterAddons{
					Addons:   k,
					Upstream: v,
				})
			}
			return ret
		}(),
		Status: pr.Status,
		BriefInstallerInfo: func() *pb.Installer {
			if p.Installer == nil {
				return nil
			}
			return &pb.Installer{
				Name:     p.Installer.Name,
				Version:  p.Installer.Version,
				RepoName: p.Installer.Repo,
			}
		}(),
	}
}

func ConvertModel2RepositoryInstallerObject(i *model.Installer) *repository.InstallerBrief {
	return &repository.InstallerBrief{
		Repo:      i.Repo,
		Name:      i.Name,
		Version:   i.Version,
		Installed: true,
	}
}
