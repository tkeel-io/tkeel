package util

import (
	v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	pb "github.com/tkeel-io/tkeel/api/plugin/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/repository"
)

func ConvertModel2PluginBriefObjectPb(p *model.Plugin, tenantID string) *pb.PluginBrief {
	return &pb.PluginBrief{
		Id:                p.ID,
		Version:           p.PluginVersion,
		TkeelVersion:      p.TkeelVersion,
		RegisterTimestamp: p.RegisterTimestamp,
		InstallerBrief: func() *pb.Installer {
			if p.Installer == nil {
				return nil
			}
			return &pb.Installer{
				Name:    p.Installer.Name,
				Version: p.Installer.Version,
				Repo:    p.Installer.Repo,
				Icon:    p.Installer.Icon,
				Desc:    p.Installer.Desc,
			}
		}(),
		TenantEnable: func() bool {
			for _, v := range p.EnableTenantes {
				if v.TenantID == tenantID {
					return true
				}
			}
			return false
		}(),
		Status: func() v1.PluginStatus {
			return p.Status
		}(),
	}
}

func ConvertModel2PluginObjectPb(p *model.Plugin, pr *model.PluginRoute, tenantID string) *pb.PluginObject {
	return &pb.PluginObject{
		Plugin:            ConvertModel2PluginBriefObjectPb(p, tenantID),
		AddonsPoint:       p.AddonsPoint,
		ImplementedPlugin: p.ImplementedPlugin,
		Secret:            p.Secret,
		EnableTenantes: func() []*pb.EnabledTenant {
			ret := make([]*pb.EnabledTenant, 0, len(p.EnableTenantes))
			for _, v := range p.EnableTenantes {
				if tenantID == model.TKeelTenant || tenantID == v.TenantID {
					ret = append(ret, &pb.EnabledTenant{
						TenantId:        v.TenantID,
						OperatorId:      v.OperatorID,
						EnableTimestamp: v.EnableTimestamp,
					})
				}
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
