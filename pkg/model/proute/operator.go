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
	"errors"

	"github.com/tkeel-io/tkeel/pkg/model"
)

var (
	ErrPluginRouteExsist          = errors.New("error plugin route existed")
	ErrPluginRouteNotExsist       = errors.New("error plugin route not existed")
	ErrPluginRouteVersionMismatch = errors.New("error plugin route version mismatch")
)

// Operator contains all operations to plugin route.
// Version.
type Operator interface {
	// Create plugin route.
	Create(context.Context, *model.PluginRoute) error
	// Update plugin route.
	Update(context.Context, *model.PluginRoute) error
	// Get plugin route with the pluginID.
	Get(ctx context.Context, pluginID string) (*model.PluginRoute, error)
	// Delete plugin route with the pluginID.
	Delete(ctx context.Context, pluginID string) (*model.PluginRoute, error)
	// Watch plugin proxy route map change.
	Watch(ctx context.Context, interval string, callback func(model.PluginProxyRouteMap) error) error
}
