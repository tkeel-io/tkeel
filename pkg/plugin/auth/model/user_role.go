package model

type UserRole struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	RoleID   string `json:"role_id"`
}
