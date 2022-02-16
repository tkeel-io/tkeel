package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/casbin/casbin/v2"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	s_model "github.com/tkeel-io/security/model"
	pb "github.com/tkeel-io/tkeel/api/rbac/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/kv"
	"github.com/tkeel-io/tkeel/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RBACService struct {
	kvOp   kv.Operator
	rbacOp *casbin.SyncedEnforcer
	db     *gorm.DB
	pb.UnimplementedRBACServer
}

func NewRBACService(kvOp kv.Operator, db *gorm.DB, rbac *casbin.SyncedEnforcer) *RBACService {
	return &RBACService{
		db:     db,
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
	newRole := &s_model.Role{}
	exist, err := newRole.IsExisted(s.db, map[string]interface{}{"name": req.Name, "tenant_id": u.Tenant})
	if err != nil {
		log.Errorf("error role(%s/%s) exist: %s", req.Name, u.Tenant, err)
		return nil, pb.ErrInternalStore()
	}
	if exist {
		return nil, pb.ErrRoleHasBeenExsist()
	}
	addPmPathSet, err := util.GetPermissionPathSet(req.Role.PermissionPathList)
	if err != nil {
		log.Errorf("error GetPermissionPathSet(%s/%s/%v)",
			req.Name, u.Tenant, req.Role.PermissionPathList)
		return nil, pb.ErrInternalStore()
	}
	rblist, err := s.addRolePermissionSet(req.Name, u.Tenant, addPmPathSet)
	if err != nil {
		log.Errorf("error add role(%s/%s/%s) permission list: %s", err)
		return nil, pb.ErrInternalStore()
	}
	defer rblist.Run()
	newRole.Name = req.Name
	newRole.TenantID = u.Tenant
	newRole.Description = req.Role.Desc
	if err = newRole.Create(s.db); err != nil {
		log.Errorf("error create role(%s): %s", newRole, err)
		return nil, pb.ErrInternalStore()
	}
	rblist = util.NewRollbackStack()
	return &pb.CreateRoleResponse{
		Role: &pb.Role{
			Name:               newRole.Name,
			Desc:               newRole.Description,
			PermissionPathList: util.Set2List(addPmPathSet),
		},
	}, nil
}

func (s *RBACService) ListRole(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.ErrInvalidArgument()
	}
	daoRole := &s_model.Role{}

	total, roles, err := daoRole.List(s.db, map[string]interface{}{"tenant_id": u.Tenant}, &s_model.Page{
		PageNum:  int(req.PageNum),
		PageSize: int(req.PageSize),
		OrderBy: func() string {
			if req.OrderBy == "" {
				return "name"
			}
			return req.OrderBy
		}(),
		IsDescending: req.IsDescending,
	}, req.KeyWords)
	if err != nil {
		log.Errorf("error role list(%s): %s", req, err)
		return nil, pb.ErrInternalStore()
	}
	return &pb.ListRolesResponse{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Total:    int32(total),
		TenantId: u.Tenant,
		Roles: func() []*pb.Role {
			ret := make([]*pb.Role, 0, len(roles))
			for _, v := range roles {
				ret = append(ret, s.convertModelRole2PB(v))
			}
			return ret
		}(),
	}, nil
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

func (s *RBACService) addRolePermissionSet(role, tenantID string, pathSet map[string]struct{}) (util.RollBackStack, error) {
	rbStack := util.NewRollbackStack()
	for _, v := range pathSet {
		if _, err := s.rbacOp.AddPolicy(role, tenantID, v, model.AllowedPermissionAction); err != nil {
			rbStack.Run()
			return nil, errors.Wrapf(err, "addRolePermissionSet add policy %s/%s/%s/%s",
				role, tenantID, v, model.AllowedPermissionAction)
		}
		rbStack = append(rbStack, func() error {
			log.Debugf("add policy %s/%s/%s/%s roll back run",
				role, tenantID, v, model.AllowedPermissionAction)
			if _, err := s.rbacOp.RemovePolicy(role, tenantID, v, model.AllowedPermissionAction); err != nil {
				return errors.Wrapf(err, "roll back remove policy %s/%s/%s/%s",
					role, tenantID, v, model.AllowedPermissionAction)
			}
			return nil
		})
	}
	return rbStack, nil
}

func (s *RBACService) convertModelRole2PB(role *s_model.Role) *pb.Role {
	ret := &pb.Role{
		Name:            role.Name,
		Desc:            role.Description,
		UpsertTimestamp: uint64(role.UpdatedAt.Unix()),
	}
	users := s.rbacOp.GetUsersForRoleInDomain(role.Name, role.TenantID)
	ret.BindNum = int32(len(users))
	// s.rbacOp.GetPermissionsForUserInDomain()
	// TODO: get role permission path.
	// s.rbacOp.AddGroupingPolicy(user,role,tenantID)
	// s.rbacOp.RemoveGroupingPolicy(user,role,tenantID)
	return ret
}
