package service

import (
	"context"

	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel/api/config/v1"
	"github.com/tkeel-io/tkeel/pkg/client/kubernetes"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ConfigService struct {
	pb.UnimplementedConfigServer

	k8s *kubernetes.Client
}

func NewConfigService(k8s *kubernetes.Client) *ConfigService {
	return &ConfigService{
		k8s: k8s,
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
