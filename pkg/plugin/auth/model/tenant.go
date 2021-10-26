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
	// TenantPrefix tenant.tenantName.
	TenantPrefix = "tenant.%s"
	// _sysTenant SysAdmin tenant.
	_sysTenant = "sys.tenant"
)

// TenantStoreOnSys tenantTitle:Tenant.
type TenantStoreOnSys map[string]*Tenant

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

func (r *Tenant) Create(ctx context.Context) error {
	var tenantStore = make(TenantStoreOnSys)
	if r.ID == "" {
		r.ID = uuid.New().String()
	}

	items, err := getDB().Select(ctx, _sysTenant)
	if err != nil {
		_dbLog.Error("[PluginAuth] Tenant Create ", err)
		return fmt.Errorf("error tenant create: %w", err)
	}
	if items != nil {
		json.Unmarshal(items, &tenantStore)
	}

	if _, ok := tenantStore[r.Title]; ok {
		return errors.New(openapi.ErrResourceExisted)
	}

	r.CreatedTime = time.Now().UTC().Unix()
	tenantStore[r.Title] = r
	saveData, _ := json.Marshal(tenantStore)

	if err := getDB().Insert(ctx, _sysTenant, saveData); err != nil {
		return fmt.Errorf("error insert: %w", err)
	}
	return nil
}

func (r *Tenant) Query(ctx context.Context) []*Tenant {
	var tenantStore = make(TenantStoreOnSys)
	items, err := getDB().Select(ctx, _sysTenant)
	if err != nil {
		_dbLog.Error("[PluginAuth] Tenant Create ", err)
		return nil
	}
	if items != nil {
		json.Unmarshal(items, &tenantStore)
	}

	var tenants = make([]*Tenant, 0)
	if tenantStore == nil {
		_dbLog.Error("please create tenant")
		return nil
	}
	if r.Title != "" {
		tenant, ok := tenantStore[r.Title]
		if ok {
			tenants = append(tenants, tenant)
			return tenants
		}
		return nil
	}

	for _, v := range tenantStore {
		tenants = append(tenants, v)
	}
	return tenants
}
