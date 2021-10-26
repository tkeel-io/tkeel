package model

import (
	"context"
	"encoding/json"
	"fmt"
)

type UserRole struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	RoleID   string `json:"role_id"`
}

const (
	// prefixUserRoleInTenant : user.role.tenantID.
	_prefixUserRoleInTenantK = "user.role.%s"
)

type (
	_emptyType          struct{}
	_roleSet            map[string]_emptyType
	_userRoleSInTenantV map[string]_roleSet
)

var _empty _emptyType

func userRoleInTenantKey(tenantID string) string {
	return fmt.Sprintf(_prefixUserRoleInTenantK, tenantID)
}

func UserRoleAdd(ctx context.Context, tenantID, userID, roleID string) error {
	data, _ := getDB().Select(ctx, userRoleInTenantKey(tenantID))
	if data == nil {
		roleSet := make(_roleSet)
		roleSet[roleID] = _empty
		userRoles := make(_userRoleSInTenantV)
		userRoles[userID] = roleSet
		userRolesData, _ := json.Marshal(userRoles)
		if err := getDB().Insert(ctx, userRoleInTenantKey(tenantID), userRolesData); err != nil {
			_dbLog.Error("model user role add", err)
			return fmt.Errorf("role role add insert %w", err)
		}
		return nil
	}
	userRoles := make(_userRoleSInTenantV)
	json.Unmarshal(data, &userRoles)
	_, ok := userRoles[userID][roleID]
	if !ok {
		userRoles[userID][roleID] = _empty
	}
	result, _ := json.Marshal(userRoles)
	if err := getDB().Insert(ctx, userRoleInTenantKey(tenantID), result); err != nil {
		_dbLog.Error("user role add insert ", err)
		return fmt.Errorf("role role add insert %w", err)
	}
	return nil
}

func UserRoleList(ctx context.Context, tenantID, userID string) (roleIDs []string) {
	data, _ := getDB().Select(ctx, userRoleInTenantKey(tenantID))
	if data == nil {
		return
	}

	userRoles := make(_userRoleSInTenantV)
	json.Unmarshal(data, &userRoles)
	roleMap, ok := userRoles[userID]
	if !ok {
		return
	}
	for i := range roleMap {
		roleIDs = append(roleIDs, i)
	}
	return
}

func UserRoleDelete(ctx context.Context, tenantID, userID, roleID string) {
	data, _ := getDB().Select(ctx, userRoleInTenantKey(tenantID))
	if data == nil {
		return
	}

	userRoles := make(_userRoleSInTenantV)
	json.Unmarshal(data, &userRoles)
	_, ok := userRoles[userID][roleID]
	if !ok {
		return
	}
	roleMap := userRoles[userID]
	delete(roleMap, roleID)
	userRoles[userID] = roleMap
	mapData, _ := json.Marshal(userRoles)
	getDB().Insert(ctx, userRoleInTenantKey(tenantID), mapData)
}
