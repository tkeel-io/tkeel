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

	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"

	dapr "github.com/dapr/go-sdk/client"
)

type Client interface {
	// v1 openapi oprator.
	Identify(ctx context.Context, sendToPluginID string) (*openapi_v1.IdentifyResponse, error)
	AddonsIdentify(ctx context.Context, sendToPluginID string, req *openapi_v1.AddonsIdentifyRequest) (*openapi_v1.AddonsIdentifyResponse, error)
	Status(ctx context.Context, sendToPluginID string) (*openapi_v1.StatusResponse, error)
	TenantBind(ctx context.Context, sendToPluginID string, req *openapi_v1.TenantBindRequst) (*openapi_v1.TenantBindResponse, error)
	TenantUnbind(ctx context.Context, sendToPluginID string, req *openapi_v1.TenantUnbindRequst) (*openapi_v1.TenantUnbindResponse, error)
}

type DaprClient struct {
	appID  string
	Client dapr.Client
}

func NewDaprClient(appID string, c dapr.Client) *DaprClient {
	return &DaprClient{
		appID:  appID,
		Client: c,
	}
}
