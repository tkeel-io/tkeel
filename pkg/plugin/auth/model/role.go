package model

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const _rolePrefix = "role.%s"

// RoleStoreOnTenant roleName：Role.
type RoleStoreOnTenant map[string]*Role

type Role struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
}

func roleKeyOnTenant(tenantID string) string {
	return fmt.Sprintf(_rolePrefix, tenantID)
}

func roleKeyOnRoleID(roleID string) string {
	return fmt.Sprintf(_rolePrefix, roleID)
}

func (r *Role) Create(ctx context.Context) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}

	items, err := getDB().Select(ctx, roleKeyOnTenant(r.TenantID))
	if err != nil {
		_dbLog.Error("[plugin auth] role create ", err)
		return fmt.Errorf("error role create: %w", err)
	}

	tenantRoleMap := make(RoleStoreOnTenant)
	if items != nil {
		json.Unmarshal(items, &tenantRoleMap)
	}

	tenantRoleMap[r.Name] = r
	saveData, _ := json.Marshal(tenantRoleMap)

	if err := getDB().Insert(ctx, roleKeyOnTenant(r.TenantID), saveData); err != nil {
		return fmt.Errorf("error insert: %w", err)
	}
	rData, _ := json.Marshal(r)
	getDB().Insert(ctx, roleKeyOnRoleID(r.ID), rData)
	return nil
}

func RoleQueryByID(ctx context.Context, roleID string) (*Role, error) {
	var role *Role
	roleData, err := getDB().Select(ctx, roleKeyOnRoleID(roleID))
	if err != nil {
		_dbLog.Error("role query by id ", err, roleID)
		return nil, fmt.Errorf("error role query: %w", err)
	}

	if err := json.Unmarshal(roleData, role); err != nil {
		_dbLog.Error("role query err", err)
		return nil, fmt.Errorf("error role query %w", err)
	}

	return role, nil
}

func RoleListOnTenant(ctx context.Context, tenantID string) ([]*Role, error) {
	var (
		roles         = make([]*Role, 0)
		tenantRoleMap = make(RoleStoreOnTenant)
	)

	if tenantID == "" {
		return nil, fmt.Errorf("nil tenant id at role list ")
	}
	roleMapData, err := getDB().Select(ctx, roleKeyOnTenant(tenantID))
	if err != nil {
		_dbLog.Error("role list err", err)
		return nil, fmt.Errorf("role list err: %w", err)
	}
	if err := json.Unmarshal(roleMapData, &tenantRoleMap); err != nil {
		return nil, fmt.Errorf("role list on tenant %w", err)
	}

	for _, v := range tenantRoleMap {
		roles = append(roles, v)
	}
	return roles, nil
}
func DeleteRoleByID(ctx context.Context, tenantID, roleID string) error {
	roleData, err := getDB().Select(ctx, roleKeyOnRoleID(roleID))
	if err != nil {
		_dbLog.Error("[plugin auth] get role ", err)
		return fmt.Errorf("error get role by id: %w", err)
	}
	role := &Role{}
	err = json.Unmarshal(roleData, role)
	if err != nil {
		_dbLog.Error("[plugin auth] unmarshal role ", err)
		return fmt.Errorf("error unmarshal role: %w", err)
	}

	items, err := getDB().Select(ctx, roleKeyOnTenant(tenantID))
	if err != nil {
		_dbLog.Error("[plugin auth] role delete by id ", err)
		return fmt.Errorf("error role delete by id: %w", err)
	}

	var tenantRoleStore = make(RoleStoreOnTenant)
	if items != nil {
		json.Unmarshal(items, &tenantRoleStore)
	}

	delete(tenantRoleStore, role.Name)
	saveData, _ := json.Marshal(tenantRoleStore)

	if err := getDB().Insert(ctx, roleKeyOnTenant(tenantID), saveData); err != nil {
		return fmt.Errorf("error insert: %w", err)
	}
	getDB().Delete(ctx, roleID)
	return nil
}
