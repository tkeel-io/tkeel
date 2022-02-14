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

package role

import (
	"context"

	"github.com/pkg/errors"

	"github.com/tkeel-io/tkeel/pkg/model"
)

var (
	ErrRoleExsist          = errors.New("error role existed")
	ErrRoleNotExsist       = errors.New("error role not existed")
	ErrRoleVersionMismatch = errors.New("error role version mismatch")
)

// Operator contains all operations to role.
type Operator interface {
	// Create role.
	Create(context.Context, *model.Role) error
	// Update role.
	Update(context.Context, *model.Role) error
	// Get role with the tenant_id role_name.
	Get(ctx context.Context, tenantID, roleName string) (*model.Role, error)
	// Delete role with the tenant_id role_name.
	Delete(ctx context.Context, tenantID, roleName string) (*model.Role, error)
	// List role.
	List(ctx context.Context, tenantID string) ([]*model.Role, error)
}
