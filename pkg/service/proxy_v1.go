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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tkeel-io/kit/log"
	t_dapr "github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	"github.com/tkeel-io/tkeel/pkg/proxy"
	v1 "github.com/tkeel-io/tkeel/pkg/proxy/v1"
)

type ProxyServiceV1 struct {
	watchInterval  string
	conf           *config.ProxyConf
	httpDaprClient t_dapr.Client
	pluginRouteOp  proute.Operator
}

func NewProxyServiceV1(interval string, conf *config.ProxyConf, client t_dapr.Client, op proute.Operator) *ProxyServiceV1 {
	return &ProxyServiceV1{
		watchInterval:  interval,
		conf:           conf,
		httpDaprClient: client,
		pluginRouteOp:  op,
	}
}

func (s *ProxyServiceV1) Watch(ctx context.Context, cb func(ppm model.PluginProxyRouteMap) error) error {
	if err := s.pluginRouteOp.Watch(ctx, s.watchInterval,
		func(pprm model.PluginProxyRouteMap) error {
			log.Debugf("pprm change: %s", pprm)
			if err := cb(pprm); err != nil {
				return fmt.Errorf("error call cb: %w", err)
			}
			return nil
		}); err != nil {
		return fmt.Errorf("error plugin route oprator watch: %w", err)
	}
	return nil
}

func (s *ProxyServiceV1) ProxyPlugin(ctx context.Context, resp http.ResponseWriter,
	req *proxy.Reqeust) error {
	log.Debugf("proxy call plugin %s", req)
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	dstResp, err := s.httpDaprClient.Call(ctxTimeout, &t_dapr.AppRequest{
		ID:         req.ID,
		Method:     req.Method,
		Verb:       req.Verb,
		Header:     req.Header,
		QueryValue: req.QueryValue,
		Body:       req.Body,
	})
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return fmt.Errorf("error plugin client call: %w", err)
	}
	if err = proxyHTTPResponse2RestfulResponse(dstResp, resp); err != nil {
		log.Errorf("error proxy http response 2 restful response: %w", err)
	}
	return nil
}

func (s *ProxyServiceV1) ProxyCore(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	log.Debugf("proxy call core %s", req.RequestURI)
	dstPath := strings.TrimPrefix(req.URL.Path, v1.ApisRootPath+v1.CoreSubPath)
	if err := proxyHTTP(ctx, s.conf.CoreAddr, dstPath, resp, req); err != nil {
		return fmt.Errorf("error proxy core: %w", err)
	}
	return nil
}

func (s *ProxyServiceV1) ProxySecurity(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	log.Debugf("proxy call security %s", req.RequestURI)
	dstPath := strings.TrimPrefix(req.URL.Path, v1.ApisRootPath+v1.SecuritySubPath)
	if err := proxyHTTP(ctx, s.conf.RudderAddr, dstPath, resp, req); err != nil {
		return fmt.Errorf("error proxy security: %w", err)
	}
	return nil
}

func (s *ProxyServiceV1) ProxyRudder(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	log.Debugf("proxy call rudder %s", req.RequestURI)
	dstPath := strings.TrimPrefix(req.URL.Path, v1.ApisRootPath+v1.RudderSubPath)
	if err := proxyHTTP(ctx, s.conf.RudderAddr, dstPath, resp, req); err != nil {
		return fmt.Errorf("error proxy rudder: %w", err)
	}
	return nil
}

func proxyHTTP(ctx context.Context, host, dstPath string,
	resp http.ResponseWriter, req *http.Request) error {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("error read request body: %w", err)
	}
	url := fmt.Sprintf("http://%s%s", host, dstPath)
	if req.URL.RawQuery != "" {
		url += "?" + req.URL.RawQuery
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	proxyReq, err := http.NewRequestWithContext(ctxTimeout, req.Method, url, bytes.NewReader(body))
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("error new proxy request: %w", err)
	}
	// proxyReq.Header = req.Header.
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}

	log.Debugf("proxy (%s --> %s)", req.URL.String(), url)
	doResp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return fmt.Errorf("error proxy call: %w", err)
	}
	if err = proxyHTTPResponse2RestfulResponse(doResp, resp); err != nil {
		return fmt.Errorf("error proxy http response 2 restful response: %w", err)
	}
	return nil
}

func proxyHTTPResponse2RestfulResponse(dstResp *http.Response, resp http.ResponseWriter) error {
	for k, vs := range dstResp.Header {
		if k == "Content-Length" {
			continue
		}
		for _, v := range vs {
			resp.Header().Add(k, v)
		}
	}
	dstBody, err := io.ReadAll(dstResp.Body)
	defer dstResp.Body.Close()
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return fmt.Errorf("error read dst response body: %w", err)
	}
	resp.WriteHeader(dstResp.StatusCode)
	if len(dstBody) == 0 {
		if _, err = resp.Write([]byte(dstResp.Status)); err != nil {
			return fmt.Errorf("error write: %w", err)
		}
	}

	var remain int
	for {
		dstBody = dstBody[remain:]
		remain, err = resp.Write(dstBody)
		if err != nil {
			log.Errorf("error write: %s", err)
		}
		if remain == 0 {
			break
		}
	}

	return nil
}
