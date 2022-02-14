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
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	pb "github.com/tkeel-io/tkeel/api/authentication/v1"
	"github.com/tkeel-io/tkeel/pkg/client"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/model"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/emicklei/go-restful"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/kit/result"
)

const (
	_securityComponent = "rudder"
	_authenticate      = "/v1/authenticate"
)

var _contextSessionKey = struct{}{}

type KeelServiceV1 struct {
	conf           *config.ProxyConf
	httpDaprClient dapr.Client
	timeout        time.Duration
}

func NewKeelServiceV1(conf *config.Configuration, client dapr.Client) *KeelServiceV1 {
	duration, err := time.ParseDuration(conf.Proxy.Timeout)
	if err != nil {
		log.Errorf("error parse duration(%s): %s", conf.Proxy.Timeout, err)
		duration = 10 * time.Second
	}
	ksV1 := &KeelServiceV1{
		conf:           conf.Proxy,
		httpDaprClient: client,
		timeout:        duration,
	}
	return ksV1
}

func writeResult(resp http.ResponseWriter, code int, msg string) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(code)
	outB, err := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}.Marshal(&result.Http{
		Code: int32(code),
		Msg:  msg,
		Data: nil,
	})
	if err != nil {
		log.Errorf("error protojson marshal: %s", err)
		resp.Write([]byte{})
		return
	}
	resp.Write(outB)
}

func (s *KeelServiceV1) Filter() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		ctx, cancel := context.WithTimeout(req.Request.Context(), s.timeout)
		defer cancel()
		sess, err := s.authenticate(ctx, req.Request)
		if err != nil {
			log.Debugf("error authenticate: %s", err)
			writeResult(resp, http.StatusUnauthorized, "authentication error")
			return
		}
		req.Request = req.Request.WithContext(withSession(ctx, sess))
		chain.ProcessFilter(req, resp)
	}
}

func (s *KeelServiceV1) ProxyPlugin(
	resp http.ResponseWriter, req *http.Request) error {
	sess, ok := getSession(req.Context())
	if !ok {
		writeResult(resp, http.StatusInternalServerError, "internal error")
		return errors.New("error get session: session not found")
	}
	log.Debugf("proxy call plugin %s", sess)
	if sess.Dst == nil {
		writeResult(resp, http.StatusInternalServerError, "internal error")
		return errors.New("error invalid dst plugin")
	}
	bodyByte := make([]byte, 0)
	if req.ContentLength != 0 {
		b, err1 := io.ReadAll(req.Body)
		if err1 != nil {
			writeResult(resp, http.StatusBadRequest, err1.Error())
			return fmt.Errorf("error read body: %w", err1)
		}
		bodyByte = b
		defer req.Body.Close()
	}

	dstResp, err := s.httpDaprClient.Call(req.Context(), &dapr.AppRequest{
		ID:         sess.Dst.ID,
		Method:     sess.RequestMethod,
		Verb:       req.Method,
		Header:     req.Header,
		QueryValue: req.URL.Query(),
		Body:       bodyByte,
	})
	if err != nil {
		writeResult(resp, http.StatusBadRequest, err.Error())
		return fmt.Errorf("error plugin client call: %w", err)
	}
	if err = proxyHTTPResponse2RestfulResponse(dstResp, resp); err != nil {
		return fmt.Errorf("error proxy http response 2 restful response: %w", err)
	}
	return nil
}

func (s *KeelServiceV1) authenticate(ctx context.Context, req *http.Request) (*session, error) {
	sess := new(session)
	out, err := s.callAuthorization(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error call authentication: %w", err)
	}
	if out.Code != http.StatusOK {
		return nil, fmt.Errorf("error call authentication: %s", out.Msg)
	}
	resp := &pb.AuthenticateResponse{}
	if err = anypb.UnmarshalTo(out.Data, resp, proto.UnmarshalOptions{}); err != nil {
		return nil, fmt.Errorf("error unmarshal resp(%s): %w", out, err)
	}
	sess.Dst = &endpoint{
		ID: resp.Destination,
	}
	u := &model.User{
		User:   resp.UserId,
		Tenant: resp.TenantId,
		Role:   resp.Role,
	}
	sess.User = u
	sess.RequestMethod = resp.Method
	req.Header.Set(model.XtKeelAuthHeader, u.Base64Encode())
	return sess, nil
}

func (s *KeelServiceV1) callAuthorization(ctx context.Context, req *http.Request) (*result.Http, error) {
	v := make(url.Values)
	v.Set("path", req.RequestURI)
	v.Set("verb", req.Method)
	out, err := client.InvokeJSON(ctx, s.httpDaprClient, &dapr.AppRequest{
		ID:         _securityComponent,
		Method:     _authenticate,
		Verb:       http.MethodGet,
		Header:     req.Header.Clone(),
		QueryValue: v,
		Body:       nil,
	}, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error invoke json: %w", err)
	}
	res := &result.Http{}
	if err = protojson.Unmarshal(out, res); err != nil {
		return nil, fmt.Errorf("error protojson unmarshal(%s): %w", out, err)
	}
	if res.Code != http.StatusOK {
		return nil, fmt.Errorf("error result: %s", res)
	}
	return res, nil
}

func withSession(ctx context.Context, sess *session) context.Context {
	return context.WithValue(ctx, _contextSessionKey, sess)
}

func getSession(ctx context.Context) (*session, bool) {
	src, ok := ctx.Value(_contextSessionKey).(*session)
	if !ok {
		return nil, false
	}
	return src, true
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
		writeResult(resp, http.StatusBadRequest, err.Error())
		return fmt.Errorf("error read dst response body: %w", err)
	}

	resp.WriteHeader(dstResp.StatusCode)
	if dstResp.ContentLength == 0 {
		return nil
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
