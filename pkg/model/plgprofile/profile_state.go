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
	"github.com/pkg/errors"
)

const (
	PRIFXITENANTPROFILEDATA = "tkeel:tenant:%s:profile:data"
	PRIFIXPROFILEPLUGIN     = "tkeel:profile:%s:plugin"
)

func profileDataKeyWithTenant(tenantID string) string {
	if tenantID == "" {
		tenantID = "_all"
	}
	return fmt.Sprintf(PRIFXITENANTPROFILEDATA, tenantID)
}

func profilePluginKey(profile string) string {
	if profile == "" {
		profile = "default"
	}
	return fmt.Sprintf(PRIFIXPROFILEPLUGIN, profile)
}

type ProfileStateStore struct {
	storeName  string
	daprClient dapr.Client
}

// nolint
func (store *ProfileStateStore) GetProfilePlugin(ctx context.Context, profile string) (plugin string, err error) {
	items, err := store.daprClient.GetState(ctx, store.storeName, profilePluginKey(profile))
	if err != nil {
		return "", err
	}

	return string(items.Value), nil
}

// nolint
func (store *ProfileStateStore) SetProfilePlugin(ctx context.Context, profile string, plugin string) error {
	return store.daprClient.SaveState(ctx, store.storeName, profilePluginKey(profile), []byte(plugin))
}

// nolint
func (store *ProfileStateStore) GetTenantProfileData(ctx context.Context, tenantID string) (data map[string]int32, err error) {
	items, err := store.daprClient.GetState(ctx, store.storeName, profileDataKeyWithTenant(tenantID))
	if err != nil {
		return nil, err
	}
	data = make(map[string]int32)
	err = json.Unmarshal(items.Value, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// nolint
func (store *ProfileStateStore) SetTenantProfileData(ctx context.Context, tenantID string, profileData map[string]int32) error {
	data, err := json.Marshal(profileData)
	if err != nil {
		return err
	}
	err = store.daprClient.SaveState(ctx, store.storeName, profileDataKeyWithTenant(tenantID), data)
	if err != nil {
		return errors.Wrapf(err, "SaveState")
	}
	return nil
}

var _ ProfileOperator = new(ProfileStateStore)

// dapr state store.
func NewProfileStateStore(storeName string, c dapr.Client) *ProfileStateStore {
	return &ProfileStateStore{
		storeName:  storeName,
		daprClient: c,
	}
}
