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

package passwd

import (
	"context"
	"fmt"

	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/model"

	dapr "github.com/dapr/go-sdk/client"
)

const (
	KeyAdminPassword = "admin_password"
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

func (o *DaprStateOprator) Create(ctx context.Context, p string) error {
	e := model.Base64Encode(p)
	log.Debugf("admin pass encode: %s", e)
	if err := o.daprClient.SaveState(ctx, o.storeName,
		KeyAdminPassword, []byte(e)); err != nil {
		return fmt.Errorf("error save admin password: %w", err)
	}
	return nil
}

func (o *DaprStateOprator) Update(ctx context.Context, p string) error {
	e := model.Base64Encode(p)
	log.Debugf("admin pass encode: %s", e)
	if err := o.daprClient.SaveState(ctx, o.storeName,
		KeyAdminPassword, []byte(e)); err != nil {
		return fmt.Errorf("error save admin password: %w", err)
	}
	return nil
}

func (o *DaprStateOprator) Get(ctx context.Context) (string, error) {
	i, err := o.daprClient.GetState(ctx, o.storeName, KeyAdminPassword)
	if err != nil {
		return "", fmt.Errorf("error get admin password: %w", err)
	}
	return string(i.Value), nil
}

func (o *DaprStateOprator) Delete(ctx context.Context) error {
	if err := o.daprClient.DeleteState(ctx, o.storeName,
		KeyAdminPassword); err != nil {
		return fmt.Errorf("error delete admin password: %w", err)
	}
	return nil
}
