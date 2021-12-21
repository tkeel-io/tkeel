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

package prepo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/repository"

	dapr "github.com/dapr/go-sdk/client"
)

const KeyPluginRepoMap = "plugin_repo_map"

type DaprStateOprator struct {
	storeName  string
	daprClient dapr.Client
	cacheRepo  *sync.Map
}

func NewDaprStateOperator(storeName string, c dapr.Client) *DaprStateOprator {
	return &DaprStateOprator{
		storeName:  storeName,
		daprClient: c,
		cacheRepo:  new(sync.Map),
	}
}

// Model2Info model.PluginRepo convert to repository.Info.
func (o *DaprStateOprator) Model2Info(p *model.PluginRepo, del bool) *repository.Info {
	if del {
		o.cacheRepo.Delete(p.Name)
	} else {
		o.cacheRepo.Store(p.Name, p)
	}
	return modelConvertInfo(p)
}

// Info2Model repository.Info convert to model.PluginRepo.
func (o *DaprStateOprator) Info2Model(i *repository.Info) *model.PluginRepo {
	mpi, ok := o.cacheRepo.LoadAndDelete(i.Name)
	if ok {
		loadMP, ok := mpi.(*model.PluginRepo)
		if !ok {
			return model.NewPluginRepo(i)
		}
		return &model.PluginRepo{
			Info:            i,
			UpsertTimestamp: loadMP.UpsertTimestamp,
			Version:         loadMP.Version,
		}
	}
	return model.NewPluginRepo(i)
}

// GetChanges compare the old and the new one and get the new, update and delete.
func (o *DaprStateOprator) GetChanges(old, curr model.PluginRepoMap) (news, updates, deletes []*model.PluginRepo) {
	news = make([]*model.PluginRepo, 0, len(curr))
	updates = make([]*model.PluginRepo, 0, len(old))
	deletes = make([]*model.PluginRepo, 0, len(old))
	for k, v := range curr {
		oldV, ok := old[k]
		if !ok {
			news = append(news, v)
		}
		if oldV.Version != v.Version {
			updates = append(updates, v)
		}
	}
	for k, v := range old {
		if _, ok := curr[k]; !ok {
			deletes = append(deletes, v)
		}
	}
	return news, updates, deletes
}

func (o *DaprStateOprator) Create(ctx context.Context, i *repository.Info) error {
	pr := o.Info2Model(i)
	// get route map.
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginRepoMap)
	if err != nil {
		return fmt.Errorf("error dapr state oprator get plugin_repo_map: %w", err)
	}
	pluginProxyMap := make(model.PluginRepoMap)
	if item.Etag != "" {
		if err = json.Unmarshal(item.Value, &pluginProxyMap); err != nil {
			return fmt.Errorf("error dapr state oprator unmarshal plugin_repo_map(%s): %w", item.Value, err)
		}
	}
	if _, ok := pluginProxyMap[pr.Name]; ok {
		return ErrPluginRepoExsist
	}
	pluginProxyMap[pr.Name] = pr
	ppmByte, err := json.Marshal(pluginProxyMap)
	if err != nil {
		return fmt.Errorf("error dapr state oprator marshal plugin_repo_map: %w", err)
	}
	// save all plugins and plugin.
	err = o.daprClient.SaveBulkState(ctx, o.storeName,
		&dapr.SetStateItem{
			Key:   KeyPluginRepoMap,
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

func (o *DaprStateOprator) Update(ctx context.Context, i *repository.Info) error {
	pr := o.Info2Model(i)
	// get route map.
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginRepoMap)
	if err != nil {
		return fmt.Errorf("error dapr state oprator get plugin_repo_map: %w", err)
	}
	pluginProxyMap := make(model.PluginRepoMap)
	if item.Etag == "" {
		return ErrPluginRepoNotExsist
	}
	err = json.Unmarshal(item.Value, &pluginProxyMap)
	if err != nil {
		return fmt.Errorf("error dapr state oprator unmarshal plugin_repo_map(%s): %w", item.Value, err)
	}
	oldPr, ok := pluginProxyMap[pr.Name]
	if !ok {
		return ErrPluginRepoNotExsist
	}
	if oldPr.Version != pr.Version {
		return ErrPluginRepoVersionMismatch
	}
	// convert model version to etag.
	vI, err := strconv.Atoi(pr.Version)
	if err != nil {
		return fmt.Errorf("error dapr state oprator strconv model version(%s): %w", pr.Version, err)
	}
	pr.Version = strconv.Itoa(vI + 1)
	pluginProxyMap[pr.Name] = pr
	valueByte, err := json.Marshal(pluginProxyMap)
	if err != nil {
		return fmt.Errorf("error dapr state oprator json marshal(%s): %w", pr, err)
	}
	err = o.daprClient.SaveBulkState(ctx, o.storeName, &dapr.SetStateItem{
		Key:   KeyPluginRepoMap,
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

func (o *DaprStateOprator) Get(ctx context.Context, name string) (*repository.Info, error) {
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginRepoMap)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator get(%s): %w", name, err)
	}
	if item.Etag == "" {
		return nil, ErrPluginRepoNotExsist
	}
	pluginRepoMap := make(model.PluginRepoMap)
	err = json.Unmarshal(item.Value, &pluginRepoMap)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator unmarshal plugin_repo_map(%s): %w", item.Value, err)
	}
	pr, ok := pluginRepoMap[name]
	if !ok {
		return nil, ErrPluginRepoNotExsist
	}
	return o.Model2Info(pr, false), nil
}

func (o *DaprStateOprator) Delete(ctx context.Context, name string) (*repository.Info, error) {
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginRepoMap)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator get(%s): %w", name, err)
	}
	if item.Etag == "" {
		return nil, ErrPluginRepoNotExsist
	}
	pluginRepoMap := make(model.PluginRepoMap)
	if err = json.Unmarshal(item.Value, &pluginRepoMap); err != nil {
		return nil, fmt.Errorf("error dapr state oprator unmarshal plugin_repo_map(%s): %w", item.Value, err)
	}
	pr, ok := pluginRepoMap[name]
	if !ok {
		return nil, ErrPluginRepoNotExsist
	}
	delete(pluginRepoMap, name)
	if len(pluginRepoMap) == 0 {
		if err = o.daprClient.DeleteState(ctx, o.storeName, KeyPluginRepoMap); err != nil {
			return nil, fmt.Errorf("error dapr state oprator delete plugin_repo_map: %w", err)
		}
	}
	valueByte, err := json.Marshal(pluginRepoMap)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator json marshal(%s): %w", pr, err)
	}
	err = o.daprClient.SaveBulkState(ctx, o.storeName, &dapr.SetStateItem{
		Key:   KeyPluginRepoMap,
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
	return o.Model2Info(pr, true), nil
}

// Watch Block waiting for plugin proxy route map changes.
// when it changes, call callback function.
func (o *DaprStateOprator) Watch(ctx context.Context, interval string, callback func(news, updates, deletes []*repository.Info) error) error {
	in, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("error dapr state oprator watch parse interval(%s): %w", interval, err)
	}
	oldMap := make(model.PluginRepoMap)
	oldTag := ""
	tick := time.NewTicker(in)
	for range tick.C {
		item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginRepoMap)
		if err != nil {
			return fmt.Errorf("error dapr state oprator watch get(%s): %w", KeyPluginRepoMap, err)
		}
		if item.Etag != oldTag {
			rMap := make(model.PluginRepoMap)
			if err = json.Unmarshal(item.Value, &rMap); err != nil {
				return fmt.Errorf("error dapr state oprator watch unmarshal(%s): %w", string(item.Value), err)
			}
			news, updates, deletes := o.GetChanges(oldMap, rMap)
			if err = callback(o.modelSli2Infos(news),
				o.modelSli2Infos(updates), o.modelSli2Infos(deletes)); err != nil {
				return fmt.Errorf("error dapr state oprator watch callback(%s): %w", rMap, err)
			}
			oldTag = item.Etag
			tick.Reset(in)
		}
	}
	return nil
}

func (o *DaprStateOprator) List(ctx context.Context) ([]*repository.Info, error) {
	// get route map.
	item, err := o.daprClient.GetState(ctx, o.storeName, KeyPluginRepoMap)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator get plugin_repo_map: %w", err)
	}
	pluginProxyMap := make(model.PluginRepoMap)
	if item.Etag == "" {
		return nil, nil
	}
	err = json.Unmarshal(item.Value, &pluginProxyMap)
	if err != nil {
		return nil, fmt.Errorf("error dapr state oprator unmarshal plugin_repo_map(%s): %w", item.Value, err)
	}
	prSli := make([]*model.PluginRepo, 0, len(pluginProxyMap))
	for _, v := range pluginProxyMap {
		prSli = append(prSli, v)
	}
	return o.modelSli2Infos(prSli), nil
}

func modelConvertInfo(pr *model.PluginRepo) *repository.Info {
	return &repository.Info{
		Name:        pr.Name,
		URL:         pr.URL,
		Annotations: pr.Annotations,
	}
}

func (o *DaprStateOprator) modelSli2Infos(prs []*model.PluginRepo) []*repository.Info {
	ret := make([]*repository.Info, 0, len(prs))
	for _, v := range prs {
		ret = append(ret, modelConvertInfo(v))
	}
	return ret
}
