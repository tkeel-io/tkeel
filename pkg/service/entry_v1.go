package service

import (
	"context"
	"sort"

	"github.com/casbin/casbin/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/tkeel-io/kit/log"
	transport_http "github.com/tkeel-io/kit/transport/http"
	"github.com/tkeel-io/security/authz/rbac"
	v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	pb "github.com/tkeel-io/tkeel/api/entry/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EntryService struct {
	pb.UnimplementedEntryServer
	pOp    plugin.Operator
	tpOp   rbac.TenantPluginMgr
	rbacOp *casbin.SyncedEnforcer
}

func NewEntryService(pOp plugin.Operator, tpOp rbac.TenantPluginMgr, rbacOp *casbin.SyncedEnforcer) *EntryService {
	return &EntryService{
		pOp:    pOp,
		tpOp:   tpOp,
		rbacOp: rbacOp,
	}
}

func (s *EntryService) GetEntries(ctx context.Context, req *emptypb.Empty) (*pb.GetEntriesResponse, error) {
	header := transport_http.HeaderFromContext(ctx)
	auths, ok := header[model.XtKeelAuthHeader]
	if !ok {
		log.Error("error get auth")
		return nil, pb.EntryErrInvalidTenant()
	}
	u := new(model.User)
	if err := u.Base64Decode(auths[0]); err != nil {
		log.Errorf("error decode auth(%s): %s", auths[0], err)
		return nil, pb.EntryErrInvalidTenant()
	}
	portal := v1.ConsolePortal_admin
	if u.User != model.TKeelUser {
		portal = v1.ConsolePortal_tenant
	}
	ret := make([]*v1.ConsoleEntry, 0)
	for _, v := range s.tpOp.ListTenantPlugins(u.Tenant) {
		allow, err := s.rbacOp.Enforce(u.User, u.Tenant, v, model.AllowedPermissionAction)
		if err != nil {
			log.Errorf("error rbac enforce(%s/%s/%s/%s): %s",
				u.User, u.Tenant, v, model.AllowedPermissionAction, v, err)
			return nil, pb.EntryErrInternalError()
		}
		if allow {
			if pluginIsTkeelComponent(v) {
				continue
			}
			p, err := s.pOp.Get(ctx, v)
			if err != nil {
				log.Errorf("error get plugin(%s): %s", v, err)
				if !errors.Is(err, plugin.ErrPluginNotExsist) {
					return nil, pb.EntryErrInternalError()
				}
				continue
			}
			ret = appendEntries(ret, p.ConsoleEntries, portal)
		}
	}

	return &pb.GetEntriesResponse{
		Entries: ret,
	}, nil
}

func appendEntries(dst, src []*v1.ConsoleEntry, portal v1.ConsolePortal) []*v1.ConsoleEntry {
	for _, v := range src {
		aP, tP := separateEntry(v)
		addEntry := aP
		if portal == v1.ConsolePortal_tenant {
			addEntry = tP
		}
		if addEntry == nil {
			continue
		}
		needMerge := false
		for _, v := range dst {
			if v.Id == addEntry.Id {
				mergeEntry(v, addEntry)
				needMerge = true
				break
			}
		}
		if !needMerge {
			dst = append(dst, addEntry)
		}
	}
	return dst
}

func separateEntry(e *v1.ConsoleEntry) (*v1.ConsoleEntry, *v1.ConsoleEntry) {
	cloneOne := proto.Clone(e)
	aP, tP := &v1.ConsoleEntry{}, &v1.ConsoleEntry{}
	proto.Merge(aP, cloneOne)
	proto.Merge(tP, cloneOne)
	if e.Children == nil {
		if e.Portal == v1.ConsolePortal_admin {
			return aP, nil
		}
		return nil, tP
	}
	aP.Children = aP.Children[0:0]
	tP.Children = tP.Children[0:0]
	for _, v := range e.Children {
		a, t := separateEntry(v)
		if a != nil {
			aP.Children = append(aP.Children, a)
		}
		if t != nil {
			tP.Children = append(tP.Children, t)
		}
	}
	if aP.Entry == "" && len(aP.Children) == 0 {
		aP = nil
	}
	if tP.Entry == "" && len(tP.Children) == 0 {
		tP = nil
	}
	return aP, tP
}

func mergeEntry(dst, src *v1.ConsoleEntry) {
	if src.Children == nil {
		return
	}
	addEntries := make([]*v1.ConsoleEntry, 0, len(src.Children))
	for _, vs := range src.Children {
		needMerge := false
		for _, vd := range dst.Children {
			if vd.Id == vs.Id {
				needMerge = true
				mergeEntry(vd, vs)
				break
			}
		}
		if !needMerge {
			addEntries = append(addEntries, vs)
		}
	}
	dst.Children = append(dst.Children, addEntries...)
	sort.Sort(entrySort(dst.Children))
}

type entrySort []*v1.ConsoleEntry

func (a entrySort) Len() int           { return len(a) }
func (a entrySort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a entrySort) Less(i, j int) bool { return a[i].Id < a[j].Id }
