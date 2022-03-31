package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel/api/config/v1"
	"github.com/tkeel-io/tkeel/pkg/client/kubernetes"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/kv"
	"github.com/tkeel-io/tkeel/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ConfigService struct {
	pb.UnimplementedConfigServer
	kvOp kv.Operator
	k8s  *kubernetes.Client
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
	}, nil
}

func (s *ConfigService) GetPlatformConfig(ctx context.Context, req *emptypb.Empty) (*pb.GetPlatformConfigResponse, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.ConfigErrInternalError()
	}
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
	var extra []byte
	if u.Tenant == model.TKeelTenant &&
		u.User == model.TKeelUser {
		e, _, err := s.getExtraData(ctx)
		if err != nil {
			log.Errorf("error get extra data: %s", err)
			return nil, pb.ConfigErrInternalError()
		}
		extra = e
	}

	return &pb.GetPlatformConfigResponse{
		AdminHost:  adminHost,
		TenantHost: tenantHost,
		Port:       conf.Port,
		Extra:      extra,
	}, nil
}

func (s *ConfigService) SetPlatformExtraConfig(ctx context.Context, req *pb.SetPlatformExtraConfigRequest) (*emptypb.Empty, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.ConfigErrInternalError()
	}
	if u.Tenant != model.TKeelTenant ||
		u.User != model.TKeelUser {
		log.Error("error not admin portal")
		return nil, pb.ConfigErrNotAdminPortal()
	}
	_, ver, err := s.getExtraData(ctx)
	if err != nil {
		log.Errorf("error get extra data: %s", err)
		return nil, pb.ConfigErrInternalError()
	}
	if err = s.setExtraData(ctx, req.Extra, ver); err != nil {
		log.Errorf("error set extra data: %s", err)
		return nil, pb.ConfigErrInternalError()
	}
	log.Debugf("set extra data(%s) succ.", req.Extra)
	return &emptypb.Empty{}, nil
}

func (s *ConfigService) getExtraData(ctx context.Context) ([]byte, string, error) {
	values, ver, err := s.kvOp.Get(ctx, model.KeyPlatExtraConfig)
	if err != nil {
		return nil, "", errors.Wrap(err, "init rudder admin password")
	}
	if ver == "" {
		return nil, ver, nil
	}
	return values, ver, nil
}

func (s *ConfigService) setExtraData(ctx context.Context, raw []byte, ver string) error {
	if ver == "" {
		if err := s.kvOp.Create(ctx, model.KeyPlatExtraConfig, raw); err != nil {
			return errors.Wrap(err, "create extra config")
		}
		return nil
	}
	if err := s.kvOp.Update(ctx, model.KeyPlatExtraConfig, raw, ver); err != nil {
		return errors.Wrap(err, "update extra config")
	}
	return nil
}
