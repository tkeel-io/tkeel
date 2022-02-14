package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/casbin/casbin/v2"
	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel/api/rbac/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/kv"
	"github.com/tkeel-io/tkeel/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RBACService struct {
	kvOp   kv.Operator
	rbacOp *casbin.SyncedEnforcer
	pb.UnimplementedRBACServer
}

func NewRBACService(kvOp kv.Operator, rbac *casbin.SyncedEnforcer) *RBACService {
	return &RBACService{
		kvOp:   kvOp,
		rbacOp: rbac,
	}
}

func (s *RBACService) CreateRoles(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.ErrInvalidArgument()
	}
	for _, v := range req.Role.PermissionPathList {
		_, err = model.GetPermissionSet().GetPermission(v)
		if err != nil {
			if errors.Is(err, model.ErrPermissionNotExist) {
				return nil, pb.ErrPermissionNotFound()
			}
			log.Errorf("error create role(%s) check permission: %s", req, err)
			return nil, pb.ErrInvalidArgument()
		}
	}

	return &pb.CreateRoleResponse{}, nil
}

func (s *RBACService) ListRole(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	return &pb.ListRolesResponse{}, nil
}

func (s *RBACService) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	return &pb.DeleteRoleResponse{}, nil
}

func (s *RBACService) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	return &pb.UpdateRoleResponse{}, nil
}

func (s *RBACService) CreateRoleBinding(ctx context.Context, req *pb.CreateRoleBindingRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *RBACService) DeleteRoleBinding(ctx context.Context, req *pb.DeleteRoleBindingRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *RBACService) ListPermissions(ctx context.Context, req *pb.ListPermissionRequest) (*pb.ListPermissionResponse, error) {
	return &pb.ListPermissionResponse{}, nil
}

func (s *RBACService) CheckRolePermission(ctx context.Context, req *pb.CheckRolePermissionRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
