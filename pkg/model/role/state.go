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

// TODO: all role.
package role

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/tkeel-io/tkeel/pkg/model"

	dapr "github.com/dapr/go-sdk/client"
)

const (
	KeyPrefixRole = "role_"
	KeyAllRole    = "all_roles"
)

type AllRoles map[string]map[string]string

func (a *AllRoles) String() string {
	b, err := json.Marshal(a)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

type DaprStateOprator struct {
	storeName  string
	daprClient dapr.Client
}

// dapr state.
func NewDaprStateOperator(storeName string, c dapr.Client) *DaprStateOprator {
	return &DaprStateOprator{
		storeName:  storeName,
		daprClient: c,
	}
}

func (o *DaprStateOprator) Create(ctx context.Context, p *model.Plugin) error {
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyAllRole)
	if err != nil {
		return fmt.Errorf("error dapr state oprator create(%s) plugin get all plugin: %w", p.ID, err)
	}
	allRs := make(AllRoles)
	if item.Etag != "" {
		if err = json.Unmarshal(item.Value, &allRs); err != nil {
			return fmt.Errorf("error dapr state oprator create(%s) plugin unmarshal all plugin(%s): %w", p.ID, item.Value, err)
		}
	}
	// check exists.
	if _, ok := allRs[p.ID]; ok {
		return ErrRoleExsist
	}
	// save all plugin map.
	allRs[p.ID] = "1"
	// conver model version 2 etag.
	vI, err := strconv.Atoi(p.Version)
	if err != nil {
		return fmt.Errorf("error dapr state oprator strconv model version(%s): %w", p.Version, err)
	}
	// marshal values.
	allPsByte, err := json.Marshal(allRs)
	if err != nil {
		return fmt.Errorf("error dapr state oprator json marshal(%s): %w", p, err)
	}
	pluginByte, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("error dapr state oprator json marshal(%s): %w", p, err)
	}
	// save all plugins and plugin.
	err = o.daprClient.SaveBulkState(ctx, o.storeName, []*dapr.SetStateItem{
		{
			Key:   KeyAllPlugin,
			Value: allPsByte,
			Etag: &dapr.ETag{
				Value: func() string {
					if item.Etag != "" {
						return item.Etag
					}
					return "0"
				}(),
			},
			Options: &dapr.StateOptions{
				Concurrency: dapr.StateConcurrencyFirstWrite,
				Consistency: dapr.StateConsistencyStrong,
			},
		},
		{
			Key:   getStoreKey(KeyPrefixPlugin, p.ID),
			Value: pluginByte,
			Etag: &dapr.ETag{
				Value: strconv.Itoa(vI - 1),
			},
			Options: &dapr.StateOptions{
				Concurrency: dapr.StateConcurrencyFirstWrite,
				Consistency: dapr.StateConsistencyStrong,
			},
		},
	}...)
	if err != nil {
		return fmt.Errorf("error dapr state oprator save(%s): %w", p, err)
	}
	return nil
}

func (o *DaprStateOprator) Update(ctx context.Context, p *model.Plugin) error {
	// convert model version 2 etag.
	old := p.Version
	vI, err := strconv.Atoi(p.Version)
	if err != nil {
		return fmt.Errorf("error dapr state oprator strconv model version(%s): %w", p.Version, err)
	}
	p.Version = strconv.Itoa(vI + 1)
	valueByte, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("error dapr state oprator json marshal(%s): %w", p, err)
	}
	err = o.daprClient.SaveBulkState(ctx, o.storeName, &dapr.SetStateItem{
		Key:   getStoreKey(KeyPrefixPlugin, p.ID),
		Value: valueByte,
		Etag: &dapr.ETag{
			Value: old,
		},
		Options: &dapr.StateOptions{
			Concurrency: dapr.StateConcurrencyFirstWrite,
			Consistency: dapr.StateConsistencyStrong,
		},
	})
	if err != nil {
		return fmt.Errorf("error dapr state oprator save(%s): %w", p, err)
	}
	return nil
}

func (o *DaprStateOprator) Get(ctx context.Context, pluginID string) (*model.Plugin, error) {
	item, err := o.daprClient.GetState(ctx, o.storeName, getStoreKey(KeyPrefixPlugin, pluginID))
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator get(%s): %w", pluginID, err)
	}
	if item.Etag == "" {
		return nil, ErrPluginNotExsist
	}
	p := &model.Plugin{}
	if err = json.Unmarshal(item.Value, p); err != nil {
		return nil, fmt.Errorf("error dapr state oprator get(%s) json unmarshal(%s): %w", item.Value, pluginID, err)
	}
	return p, nil
}

func (o *DaprStateOprator) Delete(ctx context.Context, pluginID string) (*model.Plugin, error) {
	// get all plugin map.
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyAllPlugin)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator delete(%s) plugin get all plugin: %w", pluginID, err)
	}
	allPs := make(AllPlugins)
	if item.Etag != "" {
		if err = json.Unmarshal(item.Value, &allPs); err != nil {
			return nil, fmt.Errorf("error dapr state oprator delete(%s) plugin unmarshal all plugin(%s): %w", pluginID, item.Value, err)
		}
	}
	// check exists.
	if _, ok := allPs[pluginID]; !ok {
		return nil, ErrPluginNotExsist
	}
	// delete all plugin map.
	delete(allPs, pluginID)
	// marshal values.
	allPsByte, err := json.Marshal(allPs)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator delete(%s) json marshal(%s): %w", pluginID, allPs, err)
	}
	// get plugin.
	p, err := o.Get(ctx, pluginID)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator delete get(%s) : %w", pluginID, err)
	}
	// delete plugin and update all map.
	if err = o.daprClient.ExecuteStateTransaction(ctx, o.storeName, nil, []*dapr.StateOperation{
		{
			Type: dapr.StateOperationTypeDelete,
			Item: &dapr.SetStateItem{
				Key: getStoreKey(KeyPrefixPlugin, pluginID),
			},
		},
		{
			Type: dapr.StateOperationTypeUpsert,
			Item: &dapr.SetStateItem{
				Key:   item.Key,
				Value: allPsByte,
				Etag: &dapr.ETag{
					Value: item.Etag,
				},
				Options: &dapr.StateOptions{
					Concurrency: dapr.StateConcurrencyFirstWrite,
					Consistency: dapr.StateConsistencyStrong,
				},
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("error dapr state oprator delete execute state transaction(%s/%s): %w", pluginID, allPsByte, err)
	}
	return p, nil
}

func (o *DaprStateOprator) List(ctx context.Context) ([]*model.Plugin, error) {
	// get all plugin map.
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyAllPlugin)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator list plugin get all plugin: %w", err)
	}
	allPs := make(AllPlugins)
	if item.Etag != "" {
		if err = json.Unmarshal(item.Value, &allPs); err != nil {
			return nil, fmt.Errorf("error dapr state oprator list plugin unmarshal all plugin(%s): %w", item.Value, err)
		}
	}
	if len(allPs) == 0 {
		return nil, nil
	}
	// get plugins.
	ret := make([]*model.Plugin, 0, len(allPs))
	for pluginID := range allPs {
		p, err := o.Get(ctx, pluginID)
		if err != nil {
			return nil, fmt.Errorf("error dapr state oprator list get plugin(%s): %w", pluginID, err)
		}
		ret = append(ret, p)
	}
	return ret, nil
}

func getStoreKey(prefix, pluginID string) string {
	return prefix + pluginID
}
