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

	"github.com/pkg/errors"
	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	"github.com/tkeel-io/tkeel/pkg/client"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
)

// GET identify.
func (c *DaprClient) Identify(ctx context.Context, sendToPluginID string) (*openapi_v1.IdentifyResponse, error) {
	res := &openapi_v1.IdentifyResponse{}
	_, err := client.InvokeJSON(ctx, c.c, &dapr.AppRequest{
		ID:         sendToPluginID,
		Method:     "v1/identify",
		Verb:       http.MethodGet,
		Header:     c.header.Clone(),
		QueryValue: nil,
		Body:       nil,
	}, nil, res)
	if err != nil {
		return nil, errors.Wrapf(err, "dapr invoke plugin(%s) identify", sendToPluginID)
	}
	return res, nil
}

// POST addons/identify.
func (c *DaprClient) AddonsIdentify(ctx context.Context, sendToPluginID string, req *openapi_v1.AddonsIdentifyRequest) (*openapi_v1.AddonsIdentifyResponse, error) {
	res := &openapi_v1.AddonsIdentifyResponse{}
	_, err := client.InvokeJSON(ctx, c.c, &dapr.AppRequest{
		ID:         sendToPluginID,
		Method:     "v1/addons/identify",
		Verb:       http.MethodPost,
		Header:     c.header.Clone(),
		QueryValue: nil,
		Body:       nil,
	}, req, res)
	if err != nil {
		return nil, errors.Wrapf(err, "dapr invoke plugin(%s) addons identify(%s): %w", sendToPluginID, req.String())
	}
	return res, nil
}

// GET status.
func (c *DaprClient) Status(ctx context.Context, sendToPluginID string) (*openapi_v1.StatusResponse, error) {
	res := &openapi_v1.StatusResponse{}
	_, err := client.InvokeJSON(ctx, c.c, &dapr.AppRequest{
		ID:         sendToPluginID,
		Method:     "v1/status",
		Verb:       http.MethodGet,
		Header:     c.header.Clone(),
		QueryValue: nil,
		Body:       nil,
	}, nil, res)
	if err != nil {
		return nil, errors.Wrapf(err, "dapr invoke plugin(%s) status: %w", sendToPluginID)
	}
	return res, nil
}

// POST tenant/enable.
func (c *DaprClient) TenantEnable(ctx context.Context, sendToPluginID string, req *openapi_v1.TenantEnableRequest) (*openapi_v1.TenantEnableResponse, error) {
	res := &openapi_v1.TenantEnableResponse{}
	_, err := client.InvokeJSON(ctx, c.c, &dapr.AppRequest{
		ID:         sendToPluginID,
		Method:     "v1/tenant/enable",
		Verb:       http.MethodPost,
		Header:     c.header.Clone(),
		QueryValue: nil,
		Body:       nil,
	}, req, res)
	if err != nil {
		if errors.Is(err, client.ErrMethodNotAllow) {
			_, err = client.InvokeJSON(ctx, c.c, &dapr.AppRequest{
				ID:         sendToPluginID,
				Method:     "v1/tenant/enable",
				Verb:       http.MethodGet,
				Header:     c.header.Clone(),
				QueryValue: nil,
				Body:       nil,
			}, nil, res)
			if err != nil {
				return nil, errors.Wrapf(err, "dapr invoke plugin(%s) tenant enable(%s) GET", sendToPluginID, req.String())
			}
			return res, nil
		}
		return nil, errors.Wrapf(err, "dapr invoke plugin(%s) tenant enable(%s)", sendToPluginID, req.String())
	}
	return res, nil
}

// POST tenant/disable.
func (c *DaprClient) TenantDisable(ctx context.Context, sendToPluginID string, req *openapi_v1.TenantDisableRequest) (*openapi_v1.TenantDisableResponse, error) {
	res := &openapi_v1.TenantDisableResponse{}
	_, err := client.InvokeJSON(ctx, c.c, &dapr.AppRequest{
		ID:         sendToPluginID,
		Method:     "v1/tenant/disable",
		Verb:       http.MethodPost,
		Header:     c.header.Clone(),
		QueryValue: nil,
		Body:       nil,
	}, req, res)
	if err != nil {
		if errors.Is(err, client.ErrMethodNotAllow) {
			_, err = client.InvokeJSON(ctx, c.c, &dapr.AppRequest{
				ID:         sendToPluginID,
				Method:     "v1/tenant/disable",
				Verb:       http.MethodGet,
				Header:     c.header.Clone(),
				QueryValue: nil,
				Body:       nil,
			}, nil, res)
			if err != nil {
				return nil, errors.Wrapf(err, "dapr invoke plugin(%s) tenant disable(%s) GET: %w", sendToPluginID, req.String())
			}
			return res, nil
		}
		return nil, errors.Wrapf(err, "dapr invoke plugin(%s) tenant disable(%s): %w", sendToPluginID, req.String())
	}
	return res, nil
}
