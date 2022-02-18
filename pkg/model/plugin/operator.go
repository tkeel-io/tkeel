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

package plugin

import (
	"context"

	"github.com/pkg/errors"

	"github.com/tkeel-io/tkeel/pkg/model"
)

var (
	ErrPluginExsist          = errors.New("error plugin existed")
	ErrPluginNotExsist       = errors.New("error plugin not existed")
	ErrPluginVersionMismatch = errors.New("error plugin version mismatch")
)

// Operator contains all operations to plugin.
type Operator interface {
	// Create plugin.
	Create(context.Context, *model.Plugin) error
	// Update plugin.
	Update(context.Context, *model.Plugin) error
	// Get plugin with the pluginID.
	Get(ctx context.Context, pluginID string) (*model.Plugin, error)
	// Delete plugin with the pluginID.
	Delete(ctx context.Context, pluginID string) (*model.Plugin, error)
	// List plugin.
	List(context.Context) ([]*model.Plugin, error)
}
