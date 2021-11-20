package util

import (
	pb "github.com/tkeel-io/tkeel/api/plugin/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
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
	}
}
