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
	"net/http"

	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/model"
)

type Client interface {
	// v1 openapi oprator.
	Identify(ctx context.Context, sendToPluginID string) (*openapi_v1.IdentifyResponse, error)
	AddonsIdentify(ctx context.Context, sendToPluginID string, req *openapi_v1.AddonsIdentifyRequest) (*openapi_v1.AddonsIdentifyResponse, error)
	Status(ctx context.Context, sendToPluginID string) (*openapi_v1.StatusResponse, error)
	TenantEnable(ctx context.Context, sendToPluginID string, req *openapi_v1.TenantEnableRequest) (*openapi_v1.TenantEnableResponse, error)
	TenantDisable(ctx context.Context, sendToPluginID string, req *openapi_v1.TenantDisableRequest) (*openapi_v1.TenantDisableResponse, error)
}

type DaprClient struct {
	c      *dapr.HTTPClient
	header http.Header
}

func NewDaprClient(daprHTTPPort string) *DaprClient {
	user := new(model.User)
	user.Tenant = model.TKeelTenant
	user.User = model.TKeelUser
	user.Role = model.AdminRole
	header := http.Header{}
	header.Set(model.XtKeelAuthHeader, user.Base64Encode())
	return &DaprClient{
		header: header,
		c:      dapr.NewHTTPClient(daprHTTPPort),
	}
}
