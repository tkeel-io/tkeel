package service

import (
	"context"

	"github.com/casbin/casbin/v2"
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
	ret := make([]*v1.ConsoleEntry, 0)
	for _, v := range s.tpOp.ListTenantPlugins(u.Tenant) {
		allow, err := s.rbacOp.Enforce(u.User, u.Tenant, v, model.AllowedPermissionAction)
		if err != nil {
			log.Errorf("error rbac enforce(%s/%s/%s/%s): %s",
				u.User, u.Tenant, v, model.AllowedPermissionAction, v, err)
			return nil, pb.EntryErrInternalError()
		}
		if allow {
			p, err := s.pOp.Get(ctx, v)
			if err != nil {
				if !errors.Is(err, plugin.ErrPluginNotExsist) {
					log.Errorf("error get plugin(%s): %s", v, err)
					return nil, pb.EntryErrInternalError()
				}
				continue
			}
			ret = append(ret, p.ConsoleEntries...)
		}
	}

	return &pb.GetEntriesResponse{
		Entries: ret,
	}, nil
}
