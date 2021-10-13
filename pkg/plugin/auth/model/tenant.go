package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tkeel-io/tkeel/pkg/openapi"

	"github.com/google/uuid"
)

const (
	// tenant.tenantName.
	TenantPrefix = "tenant.%s"
	// SysAdmin tenant.
	SysTenant = "sys.tenant"
)

var (
	gTenantStore = make(TenantStoreOnSys)
)

// 租户.
type Tenant struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Country     string `json:"country"`
	City        string `json:"city"`
	Address     string `json:"address" `
	CreatedTime int64  `json:"created_time"`
}

// tenantTitle:Tenant.
type TenantStoreOnSys map[string]*Tenant

func (r *Tenant) Create(ctx context.Context) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}

	items, err := getDB().Select(ctx, SysTenant)
	if err != nil {
		dblog.Error("[PluginAuth] Tenant Create ", err)
		return fmt.Errorf("error tenant create: %w", err)
	}
	if items != nil {
		json.Unmarshal(items, &gTenantStore)
	}

	_, ok := gTenantStore[r.Title]
	if ok {
		return errors.New(openapi.ErrResourceExisted)
	}

	r.CreatedTime = time.Now().UTC().Unix()
	gTenantStore[r.Title] = r
	saveData, _ := json.Marshal(gTenantStore)

	if err := getDB().Insert(ctx, SysTenant, saveData); err != nil {
		return fmt.Errorf("error insert: %w", err)
	}
	return nil
}

func (r *Tenant) Query(ctx context.Context) []*Tenant {
	var tenants = make([]*Tenant, 0)
	if gTenantStore == nil {
		dblog.Error("please create tenant")
		return nil
	}
	if r.Title != "" {
		tenant, ok := gTenantStore[r.Title]
		if ok {
			tenants = append(tenants, tenant)
			return tenants
		}
		return nil
	}

	for _, v := range gTenantStore {
		tenants = append(tenants, v)
	}
	return tenants
}

// func genTenantStateKey(tenantName string) string {
// 	return fmt.Sprintf(TenantPrefix, tenantName)
// }.
