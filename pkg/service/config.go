package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	pb "github.com/tkeel-io/tkeel/api/config/v1"
	"github.com/tkeel-io/tkeel/pkg/client/kubernetes"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/kv"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type ConfigService struct {
	pb.UnimplementedConfigServer
	kvOp kv.Operator
	k8s  *kubernetes.Client
}

func ExtraConfigKey(key string) string {
	if key == "" {
		key = model.KeyPlatExtraConfig
	} else {
		key = fmt.Sprintf("%s_%s", model.KeyPlatExtraConfig, key)
	}
	return key
}

func NewConfigService(k8s *kubernetes.Client, kvOp kv.Operator) *ConfigService {
	return &ConfigService{
		k8s:  k8s,
		kvOp: kvOp,
	}
}

func (s *ConfigService) GetDeploymentConfig(ctx context.Context, req *emptypb.Empty) (*pb.GetDeploymentConfigResponse, error) {
	conf, err := s.k8s.GetDeploymentConfig(ctx)
	if err != nil {
		log.Errorf("error get deployment config: %s", err)
		return nil, pb.ConfigErrInternalError()
	}
	adminHost, tenantHost := "", ""
	if conf.Host != nil {
		adminHost = conf.Host.Admin
		tenantHost = conf.Host.Tenant
	}

	return &pb.GetDeploymentConfigResponse{
		AdminHost:  adminHost,
		TenantHost: tenantHost,
		Port:       conf.Port,
		DocsAddr: func() string {
			if conf.Port != "80" {
				return tenantHost + ":" + conf.Port + "/docs"
			}
			return tenantHost + "/docs"
		}(),
	}, nil
}

func (s *ConfigService) GetPlatformConfig(ctx context.Context, req *pb.PlatformConfigRequest) (*structpb.Value, error) {
	key := req.Key
	path := req.Path
	extData, _, err := s.getExtraData(ctx, key)
	if err != nil {
		log.Errorf("error get extra data: %s", err)
		return nil, pb.ConfigErrInternalError()
	}
	ret := extData.Get(path)
	return NewStructValue(ret), nil
}

func (s *ConfigService) DelPlatformConfig(ctx context.Context, req *pb.PlatformConfigRequest) (*structpb.Value, error) {
	//u, err := util.GetUser(ctx)
	//if err != nil {
	//	log.Errorf("error get user: %s", err)
	//	return nil, pb.ConfigErrInternalError()
	//}
	//if u.Tenant != model.TKeelTenant ||
	//	u.User != model.TKeelUser {
	//	log.Error("error not admin portal")
	//	return nil, pb.ConfigErrNotAdminPortal()
	//}
	key := req.Key
	path := req.Path
	extData, ver, err := s.getExtraData(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get old extra data error:%w", err)
	}

	value := extData.Get(path)
	extData.Del(path)
	if err = s.setExtraData(ctx, key, extData, ver); err != nil {
		log.Errorf("error set extra data: %s", err)
		return nil, pb.ConfigErrInternalError()
	}
	return NewStructValue(value), nil
}

func (s *ConfigService) SetPlatformExtraConfig(ctx context.Context, req *pb.SetPlatformExtraConfigRequest) (*structpb.Value, error) {
	//u, err := util.GetUser(ctx)
	//if err != nil {
	//	log.Errorf("error get user: %s", err)
	//	return nil, pb.ConfigErrInternalError()
	//}
	//if u.Tenant != model.TKeelTenant ||
	//	u.User != model.TKeelUser {
	//	log.Error("error not admin portal")
	//	return nil, pb.ConfigErrNotAdminPortal()
	//}
	key := req.Key
	path := req.Path
	value, err := NewCollectValue(req.Extra)
	if err != nil {
		log.Errorf("error new collect value: %s", err)
		return nil, pb.ConfigErrInternalError()
	}
	extData, ver, err := s.getExtraData(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get old extra data error:%w", err)
	}

	extData.Set(path, value)
	if err = s.setExtraData(ctx, key, extData, ver); err != nil {
		log.Errorf("error set extra data: %s", err)
		return nil, pb.ConfigErrInternalError()
	}
	return NewStructValue(value), nil
}

func (s *ConfigService) getExtraData(ctx context.Context, key string) (*tdtl.Collect, string, error) {
	values, ver, err := s.kvOp.Get(ctx, ExtraConfigKey(key))
	if err != nil {
		return nil, ver, errors.Wrap(err, "init rudder admin password")
	}
	return tdtl.New(values), ver, nil
}

func (s *ConfigService) setExtraData(ctx context.Context, key string, value tdtl.Node, version string) error {
	if err := s.kvOp.Update(ctx, ExtraConfigKey(key), value.Raw(), version); err != nil {
		return errors.Wrap(err, "update extra config")
	}
	return nil
}

func NewCollectValue(val *structpb.Value) (*tdtl.Collect, error) {
	byt, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	return tdtl.New(byt), nil
}

func NewStructValue(cc *tdtl.Collect) *structpb.Value {
	switch cc.Type() {
	case tdtl.Bool:
		ret := cc.To(tdtl.Bool)
		switch ret := ret.(type) {
		case tdtl.BoolNode:
			return structpb.NewBoolValue(bool(ret))
		}
	case tdtl.Int, tdtl.Float, tdtl.Number:
		ret := cc.To(tdtl.Number)
		switch ret := ret.(type) {
		case tdtl.IntNode:
			return structpb.NewNumberValue(float64(ret))
		case tdtl.FloatNode:
			return structpb.NewNumberValue(float64(ret))
		}
	case tdtl.String:
		return structpb.NewStringValue(cc.String())
	case tdtl.JSON, tdtl.Object, tdtl.Array:
		ret := &structpb.Struct{}
		err := json.Unmarshal(cc.Raw(), ret)
		if err != nil {
			fmt.Println("err", err)
		}
		return structpb.NewStructValue(ret)
	}
	return structpb.NewBoolValue(false)
}
