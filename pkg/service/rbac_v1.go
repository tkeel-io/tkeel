package service

import (
	"context"
	"regexp"
	"sort"

	"gorm.io/gorm"

	"github.com/casbin/casbin/v2"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/security/authz/rbac"
	s_model "github.com/tkeel-io/security/model"
	pb "github.com/tkeel-io/tkeel/api/rbac/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RBACService struct {
	tenantPluginOp rbac.TenantPluginMgr
	rbacOp         *casbin.SyncedEnforcer
	db             *gorm.DB
	pb.UnimplementedRBACServer
}

func NewRBACService(db *gorm.DB, rbac *casbin.SyncedEnforcer, tenantPluginOp rbac.TenantPluginMgr) *RBACService {
	return &RBACService{
		tenantPluginOp: tenantPluginOp,
		db:             db,
		rbacOp:         rbac,
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
	addPmPathSet, err := util.GetPermissionPathSet(req.Role.PermissionList)
	if err != nil {
		log.Errorf("error GetPermissionPathSet(%s/%s) %s",
			req.Name, u.Tenant, req.Role.String())
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
			Name:           newRole.Name,
			Desc:           newRole.Description,
			PermissionList: util.ModelSet2PbList(addPmPathSet),
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
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.ErrInvalidArgument()
	}
	deleteRole, err := s.getDBRole(req.Name, u.Tenant)
	if err != nil {
		log.Errorf("error getDBRole(%s/%s): %s", req.Name, u.Tenant, err)
		return nil, pb.ErrInternalStore()
	}
	retPB := s.convertModelRole2PB(deleteRole)
	rbStack, err := s.deleteRoleInTenant(req.Name, u.Tenant)
	if err != nil {
		log.Errorf("error deleteRoleInTenant(%s/%s): %s", req.Name, u.Tenant, err)
		return nil, pb.ErrInternalError()
	}
	defer rbStack.Run()
	count, err := deleteRole.Delete(s.db)
	if err != nil {
		log.Errorf("error delete role(%s): %s", deleteRole, err)
		return nil, pb.ErrInternalError()
	}
	if count != 1 {
		log.Errorf("error delete role(%s): count(%d) is invalid", deleteRole, count)
		return nil, pb.ErrInternalError()
	}
	return &pb.DeleteRoleResponse{
		Role: retPB,
	}, nil
}

func (s *RBACService) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.ErrInvalidArgument()
	}
	updateRole, err := s.getDBRole(req.Name, u.Tenant)
	if err != nil {
		log.Errorf("error getDBRole(%s/%s): %s", req.Name, u.Tenant, err)
		return nil, pb.ErrInternalStore()
	}
	dbUpdateMap := setPB2Model(req.Role, updateRole)
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	if len(req.Role.PermissionList) != 0 {
		rbList, err := s.deleteRoleInTenant(req.Name, u.Tenant)
		if err != nil {
			log.Errorf("error deleteRoleInTenant(%s/%s): %s", req.Name, u.Tenant, err)
			return nil, pb.ErrInternalError()
		}
		rbStack = append(rbStack, rbList...)
		addPmPathSet, err := util.GetPermissionPathSet(req.Role.PermissionList)
		if err != nil {
			log.Errorf("error GetPermissionPathSet(%s/%s) %s",
				req.Name, u.Tenant, req.Role.String())
			return nil, pb.ErrInternalStore()
		}
		rblist, err := s.addRolePermissionSet(req.Name, u.Tenant, addPmPathSet)
		if err != nil {
			log.Errorf("error add role(%s/%s/%s) permission list: %s", err)
			return nil, pb.ErrInternalStore()
		}
		rbStack = append(rbStack, rblist...)
	}
	if len(dbUpdateMap) != 0 {
		count, err := updateRole.Update(s.db,
			map[string]interface{}{"name": updateRole.Name, "tenant_id": updateRole.TenantID},
			dbUpdateMap)
		if err != nil {
			log.Errorf("error update role(%s/%s): %s",
				updateRole.Name, updateRole.TenantID, err)
			return nil, pb.ErrInternalStore()
		}
		if count != 1 {
			log.Errorf("error update role(%s/%s): count(%d) is invalid",
				updateRole.Name, updateRole.TenantID, count)
		}
	}
	rbStack = util.NewRollbackStack()
	return &pb.UpdateRoleResponse{}, nil
}

func (s *RBACService) CreateRoleBinding(ctx context.Context, req *pb.CreateRoleBindingRequest) (*emptypb.Empty, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.ErrInvalidArgument()
	}
	daoRole := &s_model.Role{}
	exist, err := daoRole.IsExisted(s.db, map[string]interface{}{"name": req.RoleName, "tenant_id": u.Tenant})
	if err != nil {
		log.Errorf("error role(%s/%s) exist: %s", req.RoleName, u.Tenant, err)
		return nil, pb.ErrInternalStore()
	}
	if !exist {
		log.Errorf("error role(%s/%s): not found", req.RoleName, u.Tenant)
		return nil, pb.ErrRoleNotFound()
	}
	daoUser := &s_model.User{
		ID: req.User.Id,
	}
	exist, err = daoUser.Existed(s.db)
	if err != nil {
		log.Errorf("error user(%s/%s) exist: %s", req.User.Id, u.Tenant, err)
		return nil, pb.ErrInternalStore()
	}
	if !exist {
		log.Errorf("error user(%s/%s): not found", req.User.Id, u.Tenant)
		return nil, pb.ErrUserNotFound()
	}
	if _, err = s.rbacOp.AddGroupingPolicy(req.User.Id, req.RoleName, u.Tenant); err != nil {
		log.Errorf("error AddGroupingPolicy(%s/%s/%s): %s", req.User.Id, req.RoleName, u.Tenant, err)
		return nil, pb.ErrInternalStore()
	}
	return &emptypb.Empty{}, nil
}

func (s *RBACService) DeleteRoleBinding(ctx context.Context, req *pb.DeleteRoleBindingRequest) (*emptypb.Empty, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.ErrInvalidArgument()
	}

	exist := false
	roles := s.rbacOp.GetRolesForUserInDomain(req.UserId, u.Tenant)
	for _, v := range roles {
		if v == req.RoleName {
			exist = true
			break
		}
	}
	if !exist {
		log.Errorf("error DeleteRoleBinding req(%s): not exist", req)
		return nil, pb.ErrRoleNotFound()
	}
	if _, err = s.rbacOp.RemoveGroupingPolicy(req.UserId, req.RoleName, u.Tenant); err != nil {
		log.Errorf("error RemoveGroupingPolicy(%s/%s/%s): %s",
			req.UserId, req.RoleName, u.Tenant, err)
		return nil, pb.ErrRoleNotFound()
	}

	return &emptypb.Empty{}, nil
}

func (s *RBACService) ListPermissions(ctx context.Context, req *pb.ListPermissionRequest) (*pb.ListPermissionResponse, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.ErrInvalidArgument()
	}
	pList := make([]*model.Permission, 0)
	if req.Role != "" {
		daoRole := &s_model.Role{}
		exist, err1 := daoRole.IsExisted(s.db, map[string]interface{}{"name": req.Role, "tenant_id": u.Tenant})
		if err1 != nil {
			log.Errorf("error role(%s/%s) exist: %s", req.Role, u.Tenant, err1)
			return nil, pb.ErrInternalStore()
		}
		if !exist {
			log.Errorf("error role(%s/%s): not found", req.Role, u.Tenant)
			return nil, pb.ErrRoleNotFound()
		}
		pList = s.getRolePermissions(req.Role, u.Tenant)
	} else {
		plugins := s.tenantPluginOp.ListTenantPlugins(u.Tenant)
		for _, v := range plugins {
			pluginPermissions := model.GetPermissionSet().GetPermissionByPluginID(v)
			pList = append(pList, pluginPermissions...)
		}
		if !sort.IsSorted(model.PermissionSort(pList)) {
			sort.Sort(model.PermissionSort(pList))
		}
	}

	regular := getReglarStringKeyWords(req.KeyWords)
	exp, err := regexp.Compile(regular)
	if err != nil {
		log.Errorf("error create %s/%s regular expressions: %s",
			req.KeyWords, regular, err)
		return nil, pb.ErrInternalError()
	}
	ret := make([]*pb.Permission, 0, len(pList))
	for _, v := range pList {
		if exp.MatchString(v.Pb.Name) {
			ret = append(ret, &pb.Permission{
				Path:       v.Path,
				Permission: v.Pb,
			})
		}
	}

	return &pb.ListPermissionResponse{
		Permissions: ret,
	}, nil
}

func (s *RBACService) CheckRolePermission(ctx context.Context, req *pb.CheckRolePermissionRequest) (*pb.CheckRolePermissionResponse, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.ErrInvalidArgument()
	}
	allow, err := s.rbacOp.Enforce(u.User, u.Tenant, req.Path, model.AllowedPermissionAction)
	if err != nil {
		log.Errorf("error enforce(%s/%s/%s/%s): %s",
			u.User, u.Tenant, req.Path, model.AllowedPermissionAction, err)
		return nil, pb.ErrInternalError()
	}
	return &pb.CheckRolePermissionResponse{
		Allowed: allow,
	}, nil
}

func (s *RBACService) TMAddPolicy(ctx context.Context, req *pb.TMPolicyRequest) (*emptypb.Empty, error) {
	daoRole := &s_model.Role{}
	ok, err := daoRole.IsExisted(s.db, map[string]interface{}{"name": req.Role, "tenant_id": req.Tenant})
	if err != nil {
		log.Errorf("error role exist(%s): %s", req, err)
		return nil, pb.ErrInternalError()
	}
	if !ok {
		daoRole.Name = req.Role
		daoRole.TenantID = req.Tenant
		if err = daoRole.Create(s.db); err != nil {
			log.Errorf("error create role(%s): %s", req, err)
			return nil, pb.ErrInternalError()
		}
	}
	if _, err = s.rbacOp.AddPolicy(req.Role, req.Tenant,
		req.Permission, model.AllowedPermissionAction); err != nil {
		log.Errorf("error AddPolicy add policy %s/%s/%s/%s: %s",
			req.Role, req.Tenant, req.Permission, model.AllowedPermissionAction, err)
		return nil, pb.ErrInternalError()
	}
	return &emptypb.Empty{}, nil
}

func (s *RBACService) TMDeletePolicy(ctx context.Context, req *pb.TMPolicyRequest) (*emptypb.Empty, error) {
	if _, err := s.rbacOp.RemovePolicy(req.Role, req.Tenant,
		req.Permission, model.AllowedPermissionAction); err != nil {
		log.Errorf("error RemovePolicy add policy %s/%s/%s/%s: %s",
			req.Role, req.Tenant, req.Permission, model.AllowedPermissionAction, err)
		return nil, pb.ErrInternalError()
	}
	return &emptypb.Empty{}, nil
}

func (s *RBACService) TMAddRoleBinding(ctx context.Context, req *pb.TMRoleBindingRequest) (*emptypb.Empty, error) {
	if _, err := s.rbacOp.AddGroupingPolicy(req.User, req.Role, req.Tenant); err != nil {
		log.Errorf("error AddGroupingPolicy(%s/%s/%s): %s", req.User, req.Role, req.Tenant, err)
		return nil, pb.ErrInternalStore()
	}
	return &emptypb.Empty{}, nil
}

func (s *RBACService) TMDeleteRoleBinding(ctx context.Context, req *pb.TMRoleBindingRequest) (*emptypb.Empty, error) {
	if _, err := s.rbacOp.RemoveGroupingPolicy(req.User, req.Role, req.Tenant); err != nil {
		log.Errorf("error RemoveGroupingPolicy(%s/%s/%s): %s", req.User, req.Role, req.Tenant, err)
		return nil, pb.ErrInternalStore()
	}
	return &emptypb.Empty{}, nil
}

func (s *RBACService) addRolePermissionSet(role, tenantID string, pathSet map[string]*model.Permission) (util.RollBackStack, error) {
	rbStack := util.NewRollbackStack()
	for _, v := range pathSet {
		if _, err := s.rbacOp.AddPolicy(role, tenantID, v.Path, model.AllowedPermissionAction); err != nil {
			rbStack.Run()
			return nil, errors.Wrapf(err, "addRolePermissionSet add policy %s/%s/%s/%s",
				role, tenantID, v.Path, model.AllowedPermissionAction)
		}
		rbStack = append(rbStack, func() error {
			log.Debugf("add policy %s/%s/%s/%s roll back run",
				role, tenantID, v, model.AllowedPermissionAction)
			if _, err := s.rbacOp.RemovePolicy(role, tenantID, v.Path, model.AllowedPermissionAction); err != nil {
				return errors.Wrapf(err, "roll back remove policy %s/%s/%s/%s",
					role, tenantID, v.Path, model.AllowedPermissionAction)
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
	pList := s.getRolePermissions(role.Name, role.TenantID)
	ret.PermissionList = util.ModelList2PbList(pList)
	return ret
}

func (s *RBACService) getRolePermissions(role, tenant string) []*model.Permission {
	policyList := s.rbacOp.GetFilteredPolicy(0, role, tenant)
	ret := make([]*model.Permission, 0, len(policyList))
	for _, v := range policyList {
		if len(v) == 4 {
			p, err := model.GetPermissionSet().GetPermission(v[2])
			if err != nil {
				log.Errorf("error GetPermission %s: %s", v, err)
				continue
			}
			ret = append(ret, p)
		}
	}
	sort.Sort(model.PermissionSort(ret))
	return ret
}

func (s *RBACService) deleteRoleInTenant(role, tenant string) (util.RollBackStack, error) {
	rbStack := util.NewRollbackStack()
	users := s.rbacOp.GetUsersForRoleInDomain(role, tenant)
	for _, v := range users {
		if _, err := s.rbacOp.DeleteRoleForUserInDomain(v, role, tenant); err != nil {
			rbStack.Run()
			return nil, errors.Wrapf(err, "DeleteRoleForUserInDomain(%s/%s/%s)", v, role, tenant)
		}
		rbStack = append(rbStack, func() error {
			log.Debugf("DeleteRoleForUserInDomain(%s/%s/%s) roll back run",
				v, role, tenant)
			if _, err := s.rbacOp.AddGroupingPolicy(v, role, tenant); err != nil {
				return errors.Wrapf(err, "AddGroupingPolicy(%s/%s/%s)", v, role, tenant)
			}
			return nil
		})
	}
	return rbStack, nil
}

func (s *RBACService) getDBRole(role, tenant string) (*s_model.Role, error) {
	daoRole := &s_model.Role{}
	count, roles, err := daoRole.List(s.db,
		map[string]interface{}{"name": role, "tenant_id": tenant}, nil, "")
	if err != nil {
		return nil, errors.Wrapf(err, "role(%s/%s) List", role, tenant)
	}
	if count != 1 && len(roles) != 1 {
		return nil, errors.Errorf("error role(%s/%s) List: count(%d/%d) is invalid",
			role, tenant, count, len(roles))
	}
	return roles[0], nil
}

func setPB2Model(pbR *pb.Role, modelR *s_model.Role) map[string]interface{} {
	updateMap := make(map[string]interface{})
	if pbR.Name != "" {
		modelR.Name = pbR.Name
		updateMap["name"] = pbR.Name
	}
	if pbR.Desc != "" {
		modelR.Description = pbR.Name
		updateMap["description"] = pbR.Desc
	}
	return updateMap
}
