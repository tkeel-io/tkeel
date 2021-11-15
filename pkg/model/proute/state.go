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

package proute

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/tkeel-io/tkeel/pkg/model"

	dapr "github.com/dapr/go-sdk/client"
)

const KeyPrefixPluginRoute = "plugin_route_"

type DaprStateOprator struct {
	storeName  string
	daprClient dapr.Client
}

func NewDaprStateOperator(storeName string, c dapr.Client) *DaprStateOprator {
	return &DaprStateOprator{
		storeName:  storeName,
		daprClient: c,
	}
}

func (o *DaprStateOprator) Create(ctx context.Context, pr *model.PluginRoute) error {
	// conver model version 2 etag.
	vI, err := strconv.Atoi(pr.Version)
	if err != nil {
		return fmt.Errorf("error dapr state oprator strconv model version(%s): %w", pr.Version, err)
	}
	pluginRouteByte, err := json.Marshal(pr)
	if err != nil {
		return fmt.Errorf("error dapr state oprator json marshal(%s): %w", pr, err)
	}
	// save all plugins and plugin.
	err = o.daprClient.SaveBulkState(ctx, o.storeName,
		&dapr.SetStateItem{
			Key:   getStoreKey(KeyPrefixPluginRoute, pr.ID),
			Value: pluginRouteByte,
			Etag: &dapr.ETag{
				Value: strconv.Itoa(vI - 1),
			},
			Options: &dapr.StateOptions{
				Concurrency: dapr.StateConcurrencyFirstWrite,
				Consistency: dapr.StateConsistencyStrong,
			},
		})
	if err != nil {
		return fmt.Errorf("error dapr state oprator save(%s): %w", pr, err)
	}
	return nil
}

func (o *DaprStateOprator) Update(ctx context.Context, pr *model.PluginRoute) error {
	// convert model version 2 etag.
	old := pr.Version
	vI, err := strconv.Atoi(pr.Version)
	if err != nil {
		return fmt.Errorf("error dapr state oprator strconv model version(%s): %w", pr.Version, err)
	}
	pr.Version = strconv.Itoa(vI + 1)
	valueByte, err := json.Marshal(pr)
	if err != nil {
		return fmt.Errorf("error dapr state oprator json marshal(%s): %w", pr, err)
	}
	err = o.daprClient.SaveBulkState(ctx, o.storeName, &dapr.SetStateItem{
		Key:   getStoreKey(KeyPrefixPluginRoute, pr.ID),
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
		return fmt.Errorf("error dapr state oprator save(%s): %w", pr, err)
	}
	return nil
}

func (o *DaprStateOprator) Get(ctx context.Context, pluginID string) (*model.PluginRoute, error) {
	item, err := o.daprClient.GetState(ctx, o.storeName, getStoreKey(KeyPrefixPluginRoute, pluginID))
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator get(%s): %w", pluginID, err)
	}
	if item.Etag == "" {
		return nil, ErrPluginRouteNotExsist
	}
	pr := &model.PluginRoute{}
	if err = json.Unmarshal(item.Value, pr); err != nil {
		return nil, fmt.Errorf("error dapr state oprator get(%s) json unmarshal(%s): %w", item.Value, pluginID, err)
	}
	return pr, nil
}

func (o *DaprStateOprator) Delete(ctx context.Context, pluginID string) (*model.PluginRoute, error) {
	// get plugin route.
	pr, err := o.Get(ctx, pluginID)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator delete get(%s): %w", pluginID, err)
	}
	if err = o.daprClient.DeleteState(ctx, o.storeName, getStoreKey(KeyPrefixPluginRoute, pluginID)); err != nil {
		return nil, fmt.Errorf("error dapr state oprator delete(%s): %w", pr, err)
	}
	return pr, nil
}

func getStoreKey(prefix, pluginID string) string {
	return prefix + pluginID
}
