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
	"time"

	"github.com/tkeel-io/tkeel/pkg/model"

	dapr "github.com/dapr/go-sdk/client"
)

const KeyPluginProxyRouteMap = "plugin_proxy_route_map"

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
	// get route map.
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginProxyRouteMap)
	if err != nil {
		return fmt.Errorf("error dapr state oprator get plugin_proxy_route_map: %w", err)
	}
	pluginProxyMap := make(model.PluginProxyRouteMap)
	if item.Etag != "" {
		if err = json.Unmarshal(item.Value, &pluginProxyMap); err != nil {
			return fmt.Errorf("error dapr state oprator unmarshal plugin_proxy_route_map(%s): %w", item.Value, err)
		}
	}
	if _, ok := pluginProxyMap[pr.ID]; ok {
		return ErrPluginRouteExsist
	}
	pluginProxyMap[pr.ID] = pr
	ppmByte, err := json.Marshal(pluginProxyMap)
	if err != nil {
		return fmt.Errorf("error dapr state oprator marshal plugin_proxy_route_map: %w", err)
	}
	// save all plugins and plugin.
	err = o.daprClient.SaveBulkState(ctx, o.storeName,
		&dapr.SetStateItem{
			Key:   KeyPluginProxyRouteMap,
			Value: ppmByte,
			Etag: &dapr.ETag{
				Value: func() string {
					if item.Etag == "" {
						return "1"
					}
					return item.Etag
				}(),
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
	// get route map.
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginProxyRouteMap)
	if err != nil {
		return fmt.Errorf("error dapr state oprator get plugin_proxy_route_map: %w", err)
	}
	pluginProxyMap := make(model.PluginProxyRouteMap)
	if item.Etag == "" {
		return ErrPluginRouteNotExsist
	}
	err = json.Unmarshal(item.Value, &pluginProxyMap)
	if err != nil {
		return fmt.Errorf("error dapr state oprator unmarshal plugin_proxy_route_map(%s): %w", item.Value, err)
	}
	oldPr, ok := pluginProxyMap[pr.ID]
	if !ok {
		return ErrPluginRouteExsist
	}
	if oldPr.Version != pr.Version {
		return ErrPluginRouteVersionMismatch
	}
	// convert model version to etag.
	vI, err := strconv.Atoi(pr.Version)
	if err != nil {
		return fmt.Errorf("error dapr state oprator strconv model version(%s): %w", pr.Version, err)
	}
	pr.Version = strconv.Itoa(vI + 1)
	pluginProxyMap[pr.ID] = pr
	valueByte, err := json.Marshal(pluginProxyMap)
	if err != nil {
		return fmt.Errorf("error dapr state oprator json marshal(%s): %w", pr, err)
	}
	err = o.daprClient.SaveBulkState(ctx, o.storeName, &dapr.SetStateItem{
		Key:   KeyPluginProxyRouteMap,
		Value: valueByte,
		Etag: &dapr.ETag{
			Value: item.Etag,
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
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginProxyRouteMap)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator get(%s): %w", pluginID, err)
	}
	if item.Etag == "" {
		return nil, ErrPluginRouteNotExsist
	}
	pluginProxyMap := make(model.PluginProxyRouteMap)
	err = json.Unmarshal(item.Value, &pluginProxyMap)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator unmarshal plugin_proxy_route_map(%s): %w", item.Value, err)
	}
	pr, ok := pluginProxyMap[pluginID]
	if !ok {
		return nil, ErrPluginRouteNotExsist
	}
	return pr, nil
}

func (o *DaprStateOprator) Delete(ctx context.Context, pluginID string) (*model.PluginRoute, error) {
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginProxyRouteMap)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator get(%s): %w", pluginID, err)
	}
	if item.Etag == "" {
		return nil, ErrPluginRouteNotExsist
	}
	pluginProxyMap := make(model.PluginProxyRouteMap)
	if err = json.Unmarshal(item.Value, &pluginProxyMap); err != nil {
		return nil, fmt.Errorf("error dapr state oprator unmarshal plugin_proxy_route_map(%s): %w", item.Value, err)
	}
	pr, ok := pluginProxyMap[pluginID]
	if !ok {
		return nil, ErrPluginRouteExsist
	}
	delete(pluginProxyMap, pluginID)
	if len(pluginProxyMap) == 0 {
		if err = o.daprClient.DeleteState(ctx, o.storeName, KeyPluginProxyRouteMap); err != nil {
			return nil, fmt.Errorf("error dapr state oprator delete plugin_proxy_route_map: %w", err)
		}
	}
	valueByte, err := json.Marshal(pluginProxyMap)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator json marshal(%s): %w", pr, err)
	}
	err = o.daprClient.SaveBulkState(ctx, o.storeName, &dapr.SetStateItem{
		Key:   KeyPluginProxyRouteMap,
		Value: valueByte,
		Etag: &dapr.ETag{
			Value: item.Etag,
		},
		Options: &dapr.StateOptions{
			Concurrency: dapr.StateConcurrencyFirstWrite,
			Consistency: dapr.StateConsistencyStrong,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator save(%s): %w", pr, err)
	}
	return pr, nil
}

// Watch Block waiting for plugin proxy route map changes.
// when it changes, call callback function.
func (o *DaprStateOprator) Watch(ctx context.Context, interval string, callback func(model.PluginProxyRouteMap) error) error {
	in, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("error dapr state oprator watch parse interval(%s): %w", interval, err)
	}
	oldTag := ""
	tick := time.NewTicker(in)
	for range tick.C {
		item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginProxyRouteMap)
		if err != nil {
			return fmt.Errorf("error dapr state oprator watch get(%s): %w", KeyPluginProxyRouteMap, err)
		}
		if item.Etag != oldTag {
			rMap := make(model.PluginProxyRouteMap)
			if err = json.Unmarshal(item.Value, &rMap); err != nil {
				return fmt.Errorf("error dapr state oprator watch unmarshal(%s): %w", string(item.Value), err)
			}
			if err = callback(rMap); err != nil {
				return fmt.Errorf("error dapr state oprator watch callback(%s): %w", rMap, err)
			}
			oldTag = item.Etag
			tick.Reset(in)
		}
	}
	return nil
}
