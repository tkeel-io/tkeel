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

	dapr "github.com/dapr/go-sdk/client"
)

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

func (o *DaprStateOprator) Create(ctx context.Context, key string, value []byte) error {
	if err := o.daprClient.SaveState(ctx, o.storeName, key, value,
		dapr.WithConcurrency(dapr.StateConcurrencyLastWrite), dapr.WithConsistency(dapr.StateConsistencyStrong)); err != nil {
		return fmt.Errorf("error save admin password: %w", err)
	}
	return nil
}

func (o *DaprStateOprator) Update(ctx context.Context, key string, value []byte, version string) error {
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
		return fmt.Errorf("error save admin password: %w", err)
	}
	return nil
}

func (o *DaprStateOprator) Get(ctx context.Context, key string) ([]byte, string, error) {
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
	return nil
}
