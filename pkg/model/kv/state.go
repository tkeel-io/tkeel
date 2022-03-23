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

package kv

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
)

type DaprStateOprator struct {
	storeName  string
	daprClient dapr.Client
	cache      *sync.Map
}

type DaprStateCache struct {
	Value   []byte
	Version string
	cb      func(value []byte, version string) error
}

// dapr state.
func NewDaprStateOperator(interval, storeName string, c dapr.Client) *DaprStateOprator {
	d := &DaprStateOprator{
		cache:      new(sync.Map),
		storeName:  storeName,
		daprClient: c,
	}
	go d.watcher(context.TODO(), interval)
	return d
}

func (o *DaprStateOprator) Create(ctx context.Context, key string, value []byte) error {
	if err := o.daprClient.SaveBulkState(ctx, o.storeName, &dapr.SetStateItem{
		Key:   key,
		Value: value,
		Etag: &dapr.ETag{
			Value: "",
		},
		Options: &dapr.StateOptions{
			Concurrency: dapr.StateConcurrencyLastWrite,
			Consistency: dapr.StateConsistencyStrong,
		},
	}); err != nil {
		return errors.Wrapf(err, "save KV(%s/%s)", key, value)
	}
	if vi, ok := o.cache.Load(key); ok {
		v, ok := vi.(*DaprStateCache)
		if !ok {
			return errors.New("cache type invalid")
		}
		v.Value = value
		v.Version = "1"
	}
	log.Debugf("create %s kv succ", key)
	return nil
}

func (o *DaprStateOprator) Update(ctx context.Context, key string, value []byte, version string) error {
	var cacheState *DaprStateCache
	vi, ok := o.cache.Load(key)
	if ok {
		v, ok := vi.(*DaprStateCache)
		if !ok {
			return errors.New("cache type invalid")
		}
		if version == "" {
			if v.Version == "" {
				version = "1"
			} else {
				version = v.Version
			}
		}
		cacheState = v
	}
	if err := o.daprClient.SaveBulkState(ctx, o.storeName, &dapr.SetStateItem{
		Key:   key,
		Value: value,
		Etag: &dapr.ETag{
			Value: version,
		},
		Options: &dapr.StateOptions{
			Concurrency: dapr.StateConcurrencyLastWrite,
			Consistency: dapr.StateConsistencyStrong,
		},
	}); err != nil {
		return errors.Wrapf(err, "save KV(%s/%s/%s)", key, value, version)
	}
	if cacheState != nil {
		log.Debugf("cache state(%s) %s", key, string(value))
		cacheState.Value = value
		vInt, err := strconv.Atoi(version)
		if err != nil {
			return errors.Wrapf(err, "atoi version(%s)", version)
		}
		cacheState.Version = strconv.Itoa(vInt + 1)
	}
	log.Debugf("update %s kv succ", key)
	return nil
}

func (o *DaprStateOprator) Get(ctx context.Context, key string) ([]byte, string, error) {
	if vi, ok := o.cache.Load(key); ok {
		v, ok := vi.(*DaprStateCache)
		if !ok {
			return nil, "", errors.New("cache type invalid")
		}
		return v.Value, v.Version, nil
	}
	i, err := o.daprClient.GetState(ctx, o.storeName, key)
	if err != nil {
		return nil, "", fmt.Errorf("error get admin password: %w", err)
	}
	return i.Value, i.Etag, nil
}

func (o *DaprStateOprator) Delete(ctx context.Context, key string) error {
	if err := o.daprClient.DeleteState(ctx, o.storeName,
		key); err != nil {
		return fmt.Errorf("error delete admin password: %w", err)
	}
	o.cache.Delete(key)
	log.Debugf("delete %s kv succ", key)
	return nil
}

func (o *DaprStateOprator) Watch(ctx context.Context, key string, cb func(value []byte, version string) error) error {
	c := &DaprStateCache{
		cb: cb,
	}
	_, ok := o.cache.LoadOrStore(key, c)
	if ok {
		return errors.New("already watching")
	}
	item, err := o.daprClient.GetState(ctx, o.storeName, key)
	if err != nil {
		log.Errorf("error dapr state oprator watch get(%s): %s", key, err)
		return errors.Wrapf(err, "get state %s", key)
	}
	if err = cb(item.Value, item.Etag); err != nil {
		return errors.Wrapf(err, "kv cache(%s/%s/%s) call back", key, item.Value, item.Etag)
	}

	log.Debugf("%s kv changed", key)
	return nil
}

func (o *DaprStateOprator) watcher(ctx context.Context, interval string) error {
	in, err := time.ParseDuration(interval)
	if err != nil {
		return errors.Wrapf(err, "dapr state oprator watch parse interval(%s)", interval)
	}
	tick := time.NewTicker(in)
	for range tick.C {
		o.cache.Range(func(key, value interface{}) bool {
			k, ok := key.(string)
			if !ok {
				log.Error("error dapr state kv cache oprator watch key type invalid")
				return true
			}
			v, ok := value.(*DaprStateCache)
			if !ok {
				log.Errorf("error dapr state kv cache oprator watch(%s) value type invalid", k)
				return true
			}
			item, err := o.daprClient.GetState(ctx, o.storeName, k)
			if err != nil {
				log.Errorf("error dapr state oprator watch get(%s): %s", k, err)
				return true
			}
			if item.Etag != v.Version {
				v.Value = item.Value
				v.Version = item.Etag
				if err = v.cb(v.Value, v.Version); err != nil {
					log.Errorf("error kv cache(%s/%s/%s) call back: %s", k, v.Value, v.Version, err)
					return true
				}
			}
			return true
		})
		tick.Reset(in)
	}
	return nil
}
