package service

import (
	"context"
	"net/http"

	pb "github.com/tkeel-io/tkeel/api/rbac/v1"
	"github.com/tkeel-io/tkeel/pkg/model"

	"github.com/casbin/casbin/v2"
	"github.com/tkeel-io/kit/log"
	transport_http "github.com/tkeel-io/kit/transport/http"
)

type RbacService struct {
	RBACOperator *casbin.SyncedEnforcer
	pb.UnimplementedRbacServer
}

func NewRbacService(rbacOperator *casbin.SyncedEnforcer) *RbacService {
	return &RbacService{RBACOperator: rbacOperator}
}

func (s *RbacService) CreateRoles(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	header := transport_http.HeaderFromContext(ctx)
	auths, ok := header[http.CanonicalHeaderKey(model.XtKeelAuthHeader)]
	if !ok {
		log.Error("error get auth")
		return nil, pb.ErrUnknown()
	}
	user := new(model.User)
	if err := user.Base64Decode(auths[0]); err != nil {
		log.Errorf("error decode auth(%s): %s", auths[0], err)
		return nil, pb.ErrUnknown()
	}
	ok, err := s.RBACOperator.AddGroupingPolicy(user.User, req.GetBody().GetRole(), req.GetTenantId())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &pb.CreateRoleResponse{}, nil
}
func (s *RbacService) ListRole(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	header := transport_http.HeaderFromContext(ctx)
	auths, ok := header[http.CanonicalHeaderKey(model.XtKeelAuthHeader)]
	if !ok {
		log.Error("error get auth")
		return nil, pb.ErrUnknown()
	}
	user := new(model.User)
	if err := user.Base64Decode(auths[0]); err != nil {
		log.Errorf("error decode auth(%s): %s", auths[0], err)
		return nil, pb.ErrUnknown()
	}
	roles := s.RBACOperator.GetRolesForUserInDomain(user.User, req.GetTenantId())
	return &pb.ListRolesResponse{Roles: roles}, nil
}
func (s *RbacService) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	header := transport_http.HeaderFromContext(ctx)
	auths, ok := header[http.CanonicalHeaderKey(model.XtKeelAuthHeader)]
	if !ok {
		log.Error("error get auth")
		return nil, pb.ErrUnknown()
	}
	user := new(model.User)
	if err := user.Base64Decode(auths[0]); err != nil {
		log.Errorf("error decode auth(%s): %s", auths[0], err)
		return nil, pb.ErrUnknown()
	}
	ok, err := s.RBACOperator.DeleteRoleForUserInDomain(user.User, req.GetRole(), req.GetTenantId())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &pb.DeleteRoleResponse{}, nil
}
func (s *RbacService) AddRolePermission(ctx context.Context, req *pb.AddRolePermissionRequest) (*pb.AddRolePermissionResponse, error) {
	ok, err := s.RBACOperator.AddPolicy(req.GetRole(), req.GetTenantId(), req.GetBody().GetPermissionObject(), req.GetBody().GetPermissionAction())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &pb.AddRolePermissionResponse{Ok: ok}, nil
}
func (s *RbacService) DeleteRolePermission(ctx context.Context, req *pb.DeleteRolePermissionRequest) (*pb.DeleteRolePermissionResponse, error) {
	ok, err := s.RBACOperator.RemovePolicy(req.GetRole(), req.GetTenantId(), req.GetPermissionObject(), req.GetPermissionAction())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &pb.DeleteRolePermissionResponse{Ok: ok}, nil
}
func (s *RbacService) AddUserRoles(ctx context.Context, req *pb.AddUserRolesRequest) (*pb.AddUserRolesResponse, error) {
	groupingPolices := make([][]string, len(req.GetBody().GetRoles())*len(req.GetBody().GetUserIds()))
	for i := range req.GetBody().GetUserIds() {
		for j := range req.GetBody().GetRoles() {
			groupingPolices[(i+1)*(j+1)-1] = []string{req.GetBody().GetUserIds()[i], req.GetBody().GetRoles()[j], req.GetTenantId()}
		}
	}
	_, err := s.RBACOperator.AddGroupingPolicies(groupingPolices)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &pb.AddUserRolesResponse{}, nil
}
func (s *RbacService) DeleteUserRole(ctx context.Context, req *pb.DeleteUserRoleRequest) (*pb.DeleteUserRoleResponse, error) {
	_, err := s.RBACOperator.DeleteRoleForUserInDomain(req.GetUserId(), req.GetRole(), req.GetTenantId())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &pb.DeleteUserRoleResponse{}, nil
}
func (s *RbacService) ListUserPermissions(ctx context.Context, req *pb.ListUserPermissionRequest) (*pb.ListUserPermissionResponse, error) {
	permissions := s.RBACOperator.GetPermissionsForUserInDomain(req.GetUserId(), req.GetTenantId())
	out := make([]*pb.ListPermissionDetail, len(permissions))
	for i := range permissions {
		permissionItem := &pb.ListPermissionDetail{
			Role:             permissions[i][0],
			PermissionObject: permissions[i][2],
			PermissionAction: permissions[i][3],
		}
		out[i] = permissionItem
	}
	return &pb.ListUserPermissionResponse{Permissions: out}, nil
}
func (s *RbacService) CheckUserPermission(ctx context.Context, req *pb.CheckUserPermissionRequest) (*pb.CheckUserPermissionResponse, error) {
	ok, err := s.RBACOperator.Enforce(req.GetBody().GetUserId(), req.GetBody().GetTenantId(), req.GetBody().GetPermissionObject(), req.GetBody().GetPermissionAction())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInternalError()
	}
	return &pb.CheckUserPermissionResponse{Allowed: ok}, nil
}
