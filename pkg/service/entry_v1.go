package service

import (
	"context"

	"github.com/tkeel-io/kit/log"
	transport_http "github.com/tkeel-io/kit/transport/http"
	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	pb "github.com/tkeel-io/tkeel/api/entry/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/kv"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EntryService struct {
	pb.UnimplementedEntryServer

	kvOp     kv.Operator
	pluginOp plugin.Operator
}

func NewEntryService(kvOp kv.Operator, pOp plugin.Operator) *EntryService {
	return &EntryService{
		kvOp:     kvOp,
		pluginOp: pOp,
	}
}

func (s *EntryService) GetEntries(ctx context.Context, req *emptypb.Empty) (*pb.GetEntriesResponse, error) {
	header := transport_http.HeaderFromContext(ctx)
	auths, ok := header[model.XtKeelAuthHeader]
	if !ok {
		log.Error("error get auth")
		return nil, pb.EntryErrInvalidTenant()
	}
	user := new(model.User)
	if err := user.Base64Decode(auths[0]); err != nil {
		log.Errorf("error decode auth(%s): %s", auths[0], err)
		return nil, pb.EntryErrInvalidTenant()
	}
	tbKey := model.GetTenantBindKey(user.Tenant)
	vsb, _, err := s.kvOp.Get(ctx, tbKey)
	if err != nil {
		log.Errorf("error get tenant(%s) bind: %s", user.Tenant, err)
		return nil, pb.EntryErrInternalError()
	}
	tbBinds := model.ParseTenantBind(vsb)
	ret := make([]*openapi_v1.ConsoleEntry, 0, len(tbBinds))
	for _, v := range tbBinds {
		p, err := s.pluginOp.Get(ctx, v)
		if err != nil {
			log.Errorf("error get plugin(%s): %s", v, err)
			return nil, pb.EntryErrInternalError()
		}
		ret = append(ret, p.ConsoleEntries...)
	}
	return &pb.GetEntriesResponse{
		Entries: ret,
	}, nil
}
