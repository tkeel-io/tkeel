package util

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	pb "github.com/tkeel-io/tkeel/api/rbac/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/kv"
	"google.golang.org/protobuf/proto"
)

func AddPluginPermissionOnSet(ctx context.Context, kv kv.Operator, pluginID string, ps []*v1.Permission) (RollBackStack, error) {
	rbStack := NewRollbackStack()
	old, err := model.GetPermissionSet().Marshall()
	if err != nil {
		return nil, errors.Wrap(err, "permission set marshal")
	}
	rootPB := model.GetPermissionSet().NewPluginAllowedPermission(pluginID).Pb
	rootPB.Children = ps
	_, err = model.GetPermissionSet().Add(pluginID, rootPB)
	if err != nil {
		return nil, errors.Wrapf(err, "permission set add(%s/%s)", pluginID, rootPB)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("add permission set roll back run")
		model.GetPermissionSet().Delete(pluginID)
		return nil
	})
	b, err := model.GetPermissionSet().Marshall()
	if err != nil {
		rbStack.Run()
		return nil, errors.Wrap(err, "permission set marshal")
	}
	if err = kv.Update(ctx, model.KeyPermissionSet, b, ""); err != nil {
		rbStack.Run()
		return nil, errors.Wrapf(err, "KV operator update(%s/%s)", model.KeyPermissionSet, b)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("kv add permission set roll back run")
		if err = kv.Delete(ctx, model.KeyPermissionSet); err != nil {
			return errors.Wrapf(err, "KV delete %s", model.KeyPermissionSet)
		}
		if err = kv.Create(ctx, model.KeyPermissionSet, old); err != nil {
			return errors.Wrapf(err, "KV create %s/%s", model.KeyPermissionSet, old)
		}
		return nil
	})
	return rbStack, nil
}

func DeletePluginPermissionOnSet(ctx context.Context, kv kv.Operator, pluginID string) (RollBackStack, error) {
	rbStack := NewRollbackStack()
	old, err := model.GetPermissionSet().Marshall()
	if err != nil {
		return nil, errors.Wrap(err, "permission set marshal")
	}
	model.GetPermissionSet().Delete(pluginID)
	b, err := model.GetPermissionSet().Marshall()
	if err != nil {
		return nil, errors.Wrap(err, "permission set marshal")
	}
	rbStack = append(rbStack, func() error {
		model.GetPermissionSet().Unmarshal(old)
		return nil
	})
	if err = kv.Update(ctx, model.KeyPermissionSet, b, ""); err != nil {
		rbStack.Run()
		return nil, errors.Wrapf(err, "KV operator update(%s/%s)", model.KeyPermissionSet, b)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("kv delete permission set roll back run")
		if err = kv.Delete(ctx, model.KeyPermissionSet); err != nil {
			return errors.Wrapf(err, "KV delete %s", model.KeyPermissionSet)
		}
		if err = kv.Create(ctx, model.KeyPermissionSet, old); err != nil {
			return errors.Wrapf(err, "KV create %s/%s", model.KeyPermissionSet, old)
		}
		return nil
	})
	return rbStack, nil
}

func GetPermissionAllDependence(p *v1.Permission) ([]*model.Permission, error) {
	ret := make([]*model.Permission, 0, len(p.Dependences))
	for _, v := range p.Dependences {
		p, err := model.GetPermissionSet().GetPermission(v.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "get permission by path(%s)", v.Path)
		}
		ps, err := GetPermissionAllDependence(p.Pb)
		if err != nil {
			return nil, errors.Wrapf(err, "get GetPermissionAllDependencePath(%s)", p.Path)
		}
		ret = append(ret, ps...)
	}
	return ret, nil
}

func GetPermissionPathSet(pbList []*pb.Permission) (map[string]*model.Permission, error) {
	addPmPathSet := make(map[string]*model.Permission)
	// pathList := make([]string, 0, len(pbList))
	// for _, v := range pbList {.
	// 	pLevel := strings.Split(v.Path, "/").
	// 	path := "".
	// 	for _, l := range pLevel {.
	// 		path += l.
	// 		pathList = append(pathList, path).
	// 		path += "/".
	// 	}.
	// }.
	for _, v := range pbList {
		if _, ok := addPmPathSet[v.Path]; ok {
			continue
		}
		pm, err := model.GetPermissionSet().GetPermission(v.Path)
		if err != nil {
			if errors.Is(err, model.ErrPermissionNotExist) {
				return nil, model.ErrPermissionNotExist
			}
			return nil, errors.Wrapf(err, "check permission %v", v.Path)
		}
		addPmPathSet[v.Path] = pm
		ps, err := GetPermissionAllDependence(pm.Pb)
		if err != nil {
			return nil, errors.Wrapf(err, "get permission(%s) all dependence path", pm.Path)
		}
		for _, v1 := range ps {
			addPmPathSet[v1.Path] = v1
		}
	}
	return addPmPathSet, nil
}

func ModelSet2PbList(set map[string]*model.Permission, hasChild bool) []*pb.Permission {
	ret := make([]*pb.Permission, 0, len(set))
	for _, v := range set {
		ret = append(ret, &pb.Permission{
			Path: v.Path,
			Permission: func() *v1.Permission {
				ret := &v1.Permission{}
				newPB := proto.Clone(v.Pb)
				proto.Merge(ret, newPB)
				if !hasChild {
					ret.Children = nil
				}
				return ret
			}(),
		})
	}
	return ret
}

func ModelList2PbList(list []*model.Permission, hasChild bool) []*pb.Permission {
	ret := make([]*pb.Permission, 0, len(list))
	for _, v := range list {
		ret = append(ret, &pb.Permission{
			Path: v.Path,
			Permission: func() *v1.Permission {
				ret := &v1.Permission{}
				newPB := proto.Clone(v.Pb)
				proto.Merge(ret, newPB)
				if !hasChild {
					ret.Children = nil
				}
				return ret
			}(),
		})
	}
	return ret
}
