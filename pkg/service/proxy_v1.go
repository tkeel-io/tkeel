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

package service

import (
	"context"
	"fmt"
	"net/http"

	t_dapr "github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	v1 "github.com/tkeel-io/tkeel/pkg/proxy/v1"
)

type ProxyServiceV1 struct {
	watchInterval  string
	httpDaprClient t_dapr.Client
	pluginRouteOp  proute.Operator
}

func NewProxyServiceV1(interval string, client t_dapr.Client, op proute.Operator) *ProxyServiceV1 {
	return &ProxyServiceV1{
		watchInterval:  interval,
		httpDaprClient: client,
		pluginRouteOp:  op,
	}
}

func (s *ProxyServiceV1) Watch(ctx context.Context, cb func(ppm model.PluginProxyRouteMap) error) error {
	if err := s.pluginRouteOp.Watch(ctx, s.watchInterval, cb); err != nil {
		return fmt.Errorf("error plugin route oprator watch: %w", err)
	}
	return nil
}

func (s *ProxyServiceV1) Call(ctx context.Context, req *v1.ProxyReqeust) (*http.Response, error) {
	resp, err := s.httpDaprClient.Call(ctx, &t_dapr.AppReqeust{
		ID:         req.ID,
		Method:     req.Method,
		Verb:       req.Verb,
		Header:     req.Header,
		QueryValue: req.QueryValue,
		Body:       req.Body,
	})
	if err != nil {
		return nil, fmt.Errorf("error plugin client call: %w", err)
	}
	return resp, nil
}
