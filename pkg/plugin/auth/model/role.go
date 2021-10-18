package model

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const RolePrefix = "role.%s"

var gRoleStore = make(RoleStoreOnTenant)

type Role struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// roleID：Role.
type RoleStoreOnTenant map[string]*Role

func (r *Role) Create(ctx context.Context) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}

	items, err := getDB().Select(ctx, genRoleStateKey(r.TenantID))
	if err != nil {
		dblog.Error("[PluginAuth] Role Create ", err)
		return fmt.Errorf("error role create: %w", err)
	}
	if items != nil {
		json.Unmarshal(items, &gRoleStore)
	}

	gRoleStore[r.ID] = r
	saveData, _ := json.Marshal(gRoleStore)

	if err := getDB().Insert(ctx, genRoleStateKey(r.TenantID), saveData); err != nil {
		return fmt.Errorf("error insert: %w", err)
	}
	return nil
}

func genRoleStateKey(tenantID string) string {
	return fmt.Sprintf(RolePrefix, tenantID)
}
