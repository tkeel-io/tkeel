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

package openapi

import (
	"context"
	"fmt"
	"net/http"

	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	"github.com/tkeel-io/tkeel/pkg/client"
)

// GET Identify.
func (c *DaprClient) Identify(ctx context.Context, sendToPluginID string) (*openapi_v1.IdentifyResponse, error) {
	res := &openapi_v1.IdentifyResponse{}
	err := client.DaprInvokeJSON(ctx, c.Client, sendToPluginID, "v1/identify", http.MethodGet, nil, res)
	if err != nil {
		return nil, fmt.Errorf("error dapr invoke plugin(%s) identify: %w", sendToPluginID, err)
	}
	return res, nil
}

// POST AddonsIdentify.
func (c *DaprClient) AddonsIdentify(ctx context.Context, sendToPluginID string, req *openapi_v1.AddonsIdentifyRequest) (*openapi_v1.AddonsIdentifyResponse, error) {
	res := &openapi_v1.AddonsIdentifyResponse{}
	err := client.DaprInvokeJSON(ctx, c.Client, sendToPluginID, "v1/addons/identify", http.MethodPost, req, res)
	if err != nil {
		return nil, fmt.Errorf("error dapr invoke plugin(%s) addons identify(%s): %w", sendToPluginID, req.String(), err)
	}
	return res, nil
}

// GET Status.
func (c *DaprClient) Status(ctx context.Context, sendToPluginID string) (*openapi_v1.StatusResponse, error) {
	res := &openapi_v1.StatusResponse{}
	err := client.DaprInvokeJSON(ctx, c.Client, sendToPluginID, "v1/status", http.MethodGet, nil, res)
	if err != nil {
		return nil, fmt.Errorf("error dapr invoke plugin(%s) status: %w", sendToPluginID, err)
	}
	return res, nil
}

// POST Tenantbind.
func (c *DaprClient) TenantBind(ctx context.Context, sendToPluginID string, req *openapi_v1.TenantBindRequst) (*openapi_v1.TenantBindResponse, error) {
	res := &openapi_v1.TenantBindResponse{}
	err := client.DaprInvokeJSON(ctx, c.Client, sendToPluginID, "v1/tenant/bind", http.MethodPost, req, res)
	if err != nil {
		return nil, fmt.Errorf("error dapr invoke plugin(%s) tenant bind(%s): %w", sendToPluginID, req.String(), err)
	}
	return res, nil
}

// POST TenantUnbind.
func (c *DaprClient) TenantUnbind(ctx context.Context, sendToPluginID string, req *openapi_v1.TenantUnbindRequst) (*openapi_v1.TenantUnbindResponse, error) {
	res := &openapi_v1.TenantUnbindResponse{}
	err := client.DaprInvokeJSON(ctx, c.Client, sendToPluginID, "v1/tenant/unbind", http.MethodPost, req, res)
	if err != nil {
		return nil, fmt.Errorf("error dapr invoke plugin(%s) tenant unbind(%s): %w", sendToPluginID, req.String(), err)
	}
	return res, nil
}
