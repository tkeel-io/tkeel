package util

import (
	v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
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
		Secret:            p.Secret,
		RegisterTimestamp: p.RegisterTimestamp,
		EnableTenantes: func() []*pb.EnabledTenant {
			ret := make([]*pb.EnabledTenant, 0, len(p.EnableTenantes))
			for _, v := range p.EnableTenantes {
				ret = append(ret, &pb.EnabledTenant{
					TenantId:        v.TenantID,
					OperatorId:      v.OperatorID,
					EnableTimestamp: v.EnableTimestamp,
				})
			}
			return ret
		}(),
		RegisterAddons: func() []*pb.RegisterAddons {
			if pr == nil {
				return nil
			}
			ret := make([]*pb.RegisterAddons, 0, len(pr.RegisterAddons))
			for k, v := range pr.RegisterAddons {
				ret = append(ret, &pb.RegisterAddons{
					Addons:   k,
					Upstream: v,
				})
			}
			return ret
		}(),
		Status: func() v1.PluginStatus {
			return p.Status
		}(),
		BriefInstallerInfo: func() *pb.Installer {
			if p.Installer == nil {
				return nil
			}
			return &pb.Installer{
				Name:    p.Installer.Name,
				Version: p.Installer.Version,
				Repo:    p.Installer.Repo,
			}
		}(),
		ConsoleEntries: p.ConsoleEntries,
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
