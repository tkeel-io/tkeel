package model

type RolePermission struct {
	RoleID         string `json:"role_id"`
	TenantID       string `json:"tenant_id"`
	PermissionType string `json:"permission_type"` // [PLUGIN ENTITY].
	PermissionID   string `json:"permission_id"`
}
