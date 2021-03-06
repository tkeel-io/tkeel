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

	"github.com/pkg/errors"
)

var (
	ErrKVKeyExsist            = errors.New("error Key existed")
	ErrKVKeyNotExsist         = errors.New("error Key not existed")
	ErrKVStateVersionMismatch = errors.New("error KV state version mismatch")
)

// Operator contains all operations to KV.
type Operator interface {
	// Create KV.
	Create(ctx context.Context, key string, value []byte) error
	// Update KV.
	Update(ctx context.Context, key string, value []byte, version string) error
	// Get KV.
	Get(ctx context.Context, key string) (value []byte, version string, err error)
	// Delete KV.
	Delete(ctx context.Context, key string) error
	// Watch KV.
	Watch(ctx context.Context, key string, cb func(value []byte, version string) error) error
}
