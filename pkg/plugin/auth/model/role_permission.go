package model

import (
	"context"
	"encoding/json"
	"fmt"
)

type RolePermission struct {
	RoleID         string `json:"role_id"`
	TenantID       string `json:"tenant_id"`
	PermissionType string `json:"permission_type"` // [PLUGIN ENTITY].
	PermissionID   string `json:"permission_id"`
}

const (
	// prefixUserRoleInTenant : permission.role.roleID.
	_prefixRolePermissionK = "permission.role.%s"
)

type (
	// _PermissionSet :permissionID.
	_PermissionSet map[string]_emptyType
	// _AllPermissionSetV : permissionType:permissionID.
	_AllPermissionSetV map[string]_PermissionSet
)

func rolePermissionKey(roleID string) string {
	return fmt.Sprintf(_prefixRolePermissionK, roleID)
}

func RolePermissionAdd(ctx context.Context, roleID, permissionType, permissionID string) error {
	data, _ := getDB().Select(ctx, rolePermissionKey(roleID))
	if data == nil {
		permission := make(_PermissionSet)
		permission[roleID] = _empty
		permissionSet := make(_AllPermissionSetV)
		permissionSet[permissionType] = permission
		permissionData, _ := json.Marshal(permissionSet)
		if err := getDB().Insert(ctx, rolePermissionKey(roleID), permissionData); err != nil {
			_dbLog.Error("model user role add", err)
			return fmt.Errorf("role perimission add insert %w", err)
		}
		return nil
	}
	permissionSet := make(_AllPermissionSetV)
	json.Unmarshal(data, &permissionSet)
	_, ok := permissionSet[permissionType][permissionID]
	if !ok {
		permissionSet[permissionType][permissionID] = _empty
		result, _ := json.Marshal(permissionSet)
		if err := getDB().Insert(ctx, rolePermissionKey(roleID), result); err != nil {
			_dbLog.Error("role permission  add insert ", err)
			return fmt.Errorf("role perimission add insert %w", err)
		}
		return nil
	}
	return nil
}

func RolePermissionList(ctx context.Context, permissionType, roleID string) (permissionIDs []string) {
	data, _ := getDB().Select(ctx, rolePermissionKey(roleID))
	if data == nil {
		return
	}

	permissionSet := make(_AllPermissionSetV)
	json.Unmarshal(data, &permissionSet)
	permissionMap, ok := permissionSet[permissionType]
	if !ok {
		return
	}
	for i := range permissionMap {
		permissionIDs = append(permissionIDs, i)
	}
	return
}
func RolePermissionDelete(ctx context.Context, roleID, permissionType, permissionID string) {
	data, _ := getDB().Select(ctx, rolePermissionKey(roleID))
	if data == nil {
		return
	}

	permissionSet := make(_AllPermissionSetV)
	json.Unmarshal(data, &permissionSet)
	_, ok := permissionSet[permissionType][permissionID]
	if !ok {
		return
	}
	permissionMap := permissionSet[permissionType]
	delete(permissionMap, permissionID)
	permissionSet[permissionType] = permissionMap
	mapData, _ := json.Marshal(permissionSet)
	getDB().Insert(ctx, rolePermissionKey(roleID), mapData)
}
