/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plgprofile

import (
	"context"
	"encoding/json"
	"fmt"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/tkeel-io/tkeel/pkg/model"
)

const (
	PRIFXITENANTPROFILE = "tkeel:tenant:profile:"
)

func profileKeyWithTenant(tenantID string) string {
	if tenantID == "" {
		tenantID = "_all"
	}
	return PRIFXITENANTPROFILE + tenantID
}

type ProfileStateStore struct {
	storeName  string
	daprClient dapr.Client
}

// dapr state store.
func NewProfileStateStore(storeName string, c dapr.Client) *ProfileStateStore {
	return &ProfileStateStore{
		storeName:  storeName,
		daprClient: c,
	}
}

func (store *ProfileStateStore) GetTenantProfile(ctx context.Context, tenantID string) ([]*model.PluginProfile, error) {
	profiles := make([]*model.PluginProfile, 0)
	item, err := store.daprClient.GetState(ctx, store.storeName, profileKeyWithTenant(tenantID))
	if err != nil {
		return nil, fmt.Errorf("get state %w", err)
	}
	if err = json.Unmarshal(item.Value, &profiles); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return profiles, nil
}

func (store *ProfileStateStore) SetTenantProfile(ctx context.Context, tenantID string, profile []*model.PluginProfile) error {
	items, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("set profile %w", err)
	}
	err = store.daprClient.SaveState(ctx, store.storeName, profileKeyWithTenant(tenantID), items)
	if err != nil {
		return fmt.Errorf("set profile %w", err)
	}
	return nil
}
func (store *ProfileStateStore) SetTenantPluginProfile(ctx context.Context, tenantID string, profile *model.PluginProfile) error {
	profiles := make([]*model.PluginProfile, 0)
	item, err := store.daprClient.GetState(ctx, store.storeName, profileKeyWithTenant(tenantID))
	if err != nil {
		return fmt.Errorf("get state %w", err)
	}
	if item.Value != nil {
		json.Unmarshal(item.Value, &profiles)
	}

	var update bool
	for i := range profiles {
		if profiles[i].PluginID == profile.PluginID {
			profiles[i] = profile
			update = true
			break
		}
	}
	if !update {
		profiles = append(profiles, profile)
	}

	items, err := json.Marshal(profiles)
	if err != nil {
		return fmt.Errorf("set profile %w", err)
	}
	err = store.daprClient.SaveState(ctx, store.storeName, profileKeyWithTenant(tenantID), items)
	if err != nil {
		return fmt.Errorf("set profile %w", err)
	}
	return nil
}
