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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/security/models/oauth"
	"github.com/tkeel-io/security/models/rbac"
	t_dapr "github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	v1 "github.com/tkeel-io/tkeel/pkg/service/keel/v1"
	"github.com/tkeel-io/tkeel/pkg/token"
	"github.com/tkeel-io/tkeel/pkg/util"
)

const (
	AuthorizationHeader = "Authorization"
)

var (
	contextUserKey       = struct{ K int }{K: 0}
	contextSourceKey     = struct{ K int }{K: 1}
	ErrNotFoundUpstream  = errors.New("not found upstream plugin")
	ErrNotActiveUpstream = errors.New("not active upstream plugin")
	ErrNotFoundAddons    = errors.New("not found addons")
)

type upstream struct {
	ID            string
	Method        string
	TKeelDepened  string
	ActiveTenants []string
}

func (u *upstream) Verify(req *http.Request) error {
	src, ok := getSource(req.Context())
	if !ok {
		return errors.New("invalid source")
	}
	user, ok := getUser(req.Context())
	if !ok {
		return errors.New("invalid user")
	}
	ok, err := util.CheckRegisterPluginTkeelVersion(u.TKeelDepened, src.TKeelDepened)
	if err != nil {
		return fmt.Errorf("error check depended tKeel version relationship(%s/%s): %w",
			u.TKeelDepened, src.TKeelDepened, err)
	}
	if !ok {
		return errors.New("invalid depende tKeel version")
	}
	active := false
	for _, v := range u.ActiveTenants {
		if v == user.Tenant {
			active = true
			break
		}
	}
	if !active {
		return ErrNotActiveUpstream
	}
	return nil
}

type source struct {
	ID           string
	TKeelDepened string
}

type KeelServiceV1 struct {
	watchInterval  string
	tKeelVersion   string
	conf           *config.ProxyConf
	httpDaprClient t_dapr.Client
	pluginRouteOp  proute.Operator
	pluginRouteMap *sync.Map
	secretProvider token.Provider
	timeout        time.Duration
	pathWhiteList  map[string][]string
}

func NewKeelServiceV1(interval string, conf *config.Configuration, client t_dapr.Client, op proute.Operator) *KeelServiceV1 {
	secret := util.RandStringBytesMaskImpr(0, 10)
	duration, err := time.ParseDuration(conf.Proxy.Timeout)
	if err != nil {
		log.Errorf("error parse duration(%s): %s", conf.Proxy.Timeout, err)
		duration = 10 * time.Second
	}
	ksV1 := &KeelServiceV1{
		tKeelVersion:   conf.Tkeel.Version,
		watchInterval:  interval,
		conf:           conf.Proxy,
		httpDaprClient: client,
		pluginRouteOp:  op,
		pluginRouteMap: new(sync.Map),
		secretProvider: token.InitProvider(secret, "", ""),
		timeout:        duration,
		pathWhiteList: map[string][]string{
			v1.RudderSubPath: {
				"/apis/rudder/v1/oauth2/plugin",
				"/apis/rudder/v1/oauth2/admin",
			},
			v1.SecuritySubPath: {
				"/apis/security/v1/oauth/token",
				"/apis/security/v1/oauth/authorize",
				"/apis/security/v1/oauth/authenticate",
				"/apis/security/v1/oauth/on_code",
			},
		},
	}
	if _, err := oauth.NewOperator(conf.SecurityConf.OAuth); err != nil {
		log.Fatalf("error oauth new operator: %s", err)
		return nil
	}
	if _, err := rbac.NewRBACOperator(conf.SecurityConf.Mysql); err != nil {
		log.Fatalf("error rbac new operator: %s", err)
		return nil
	}
	go func() {
		if err := ksV1.watch(context.TODO()); err != nil {
			log.Fatalf("error keel watch plugin route map: %s", err)
		}
	}()
	return ksV1
}

func (s *KeelServiceV1) watch(ctx context.Context) error {
	if err := s.pluginRouteOp.Watch(ctx, s.watchInterval,
		func(pprm model.PluginProxyRouteMap) error {
			log.Debugf("pprm change: %s", pprm)
			// upsert new route map.
			for id, v := range pprm {
				s.pluginRouteMap.Store(id, v)
			}
			// delete old route map.
			s.pluginRouteMap.Range(func(key, value interface{}) bool {
				pID, ok := key.(string)
				if !ok {
					s.pluginRouteMap.Delete(key)
					log.Errorf("error invalid key type: %v", key)
					return true
				}
				if _, ok = pprm[pID]; !ok {
					s.pluginRouteMap.Delete(key)
				}
				return true
			})
			return nil
		}); err != nil {
		return fmt.Errorf("error plugin route oprator watch: %w", err)
	}
	return nil
}

func setResult(code int, msg string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}

func writeResult(resp http.ResponseWriter, code int, msg string, data interface{}) {
	resp.WriteHeader(code)
	b, err := json.Marshal(setResult(code, msg, data))
	if err != nil {
		log.Errorf("error marshal result: %w", err)
	}
	resp.Write(b)
}

func (s *KeelServiceV1) Filter() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		// white path.
		for k, vs := range s.pathWhiteList {
			for _, v := range vs {
				if req.Request.URL.Path == v {
					log.Debugf("path(%s) in white(%s) list", req.Request.URL.Path, k)
					chain.ProcessFilter(req, resp)
					return
				}
			}
		}
		ctx, cancel := context.WithTimeout(req.Request.Context(), s.timeout)
		defer cancel()
		// with source.
		pluginID, err := s.getPluginIDFromRequest(req)
		if err != nil {
			log.Errorf("error get plugin ID from request: %s", err)
			resp.WriteHeaderAndJson(http.StatusUnauthorized,
				setResult(http.StatusUnauthorized, "invalid token", nil), "application/json")
			return
		}
		// with user.
		if pluginID == "" {
			log.Debugf("external flow")
			user, err1 := s.externalGetUser(req)
			if err1 != nil {
				log.Errorf("error external get user: %s", err1)
				resp.WriteHeaderAndJson(http.StatusUnauthorized,
					setResult(http.StatusUnauthorized, "invalid token", nil), "application/json")
				return
			}
			req.Request.Header[http.CanonicalHeaderKey(model.XtKeelAuthHeader)] = []string{user.Base64Encode()}
			ctx = withUser(ctx, user)
			ctx = withSource(ctx, &source{
				ID:           model.TKeelUser,
				TKeelDepened: s.tKeelVersion,
			})
		} else if pluginID == "rudder" || pluginID == "keel" {
			user := new(model.User)
			user.Tenant = model.TKeelTenant
			user.User = model.TKeelUser
			user.Role = model.AdminRole
			req.Request.Header[http.CanonicalHeaderKey(model.XtKeelAuthHeader)] = []string{user.Base64Encode()}
			ctx = withUser(ctx, user)
			ctx = withSource(ctx, &source{
				ID:           model.TKeelUser,
				TKeelDepened: s.tKeelVersion,
			})
		} else {
			pluginRouteInterface, ok := s.pluginRouteMap.Load(pluginID)
			if !ok {
				log.Errorf("error source plugin ID(%s) not register", pluginID)
				resp.WriteHeaderAndJson(http.StatusUnauthorized,
					setResult(http.StatusUnauthorized, "internal error", nil), "application/json")
				return
			}
			pluginRoute, ok := pluginRouteInterface.(*model.PluginRoute)
			if !ok {
				log.Error("error source plugin route type invalid")
				resp.WriteHeaderAndJson(http.StatusUnauthorized,
					setResult(http.StatusUnauthorized, "internal error", nil), "application/json")
				return
			}
			log.Debugf("internal flow")
			tKeelHeader := req.HeaderParameter(http.CanonicalHeaderKey(model.XtKeelAuthHeader))
			if tKeelHeader == "" {
				log.Errorf("error internal flow not found x-tKeel-auth")
				resp.WriteHeaderAndJson(http.StatusUnauthorized,
					setResult(http.StatusUnauthorized, "x-tKeel-auth invalid", nil), "application/json")
				return
			}
			user := new(model.User)
			if err = user.Base64Decode(tKeelHeader); err != nil {
				log.Errorf("error decode x-tKeel-auth(%s): %s", tKeelHeader, err)
				resp.WriteHeaderAndJson(http.StatusUnauthorized,
					setResult(http.StatusUnauthorized, "x-tKeel-auth invalid", nil), "application/json")
				return
			}
			ctx = withUser(ctx, user)
			ctx = withSource(ctx, &source{
				ID:           pluginID,
				TKeelDepened: pluginRoute.TkeelVersion,
			})
		}
		req.Request = req.Request.WithContext(ctx)
		chain.ProcessFilter(req, resp)
	}
}

func (s *KeelServiceV1) getPluginIDFromRequest(req *restful.Request) (string, error) {
	pluginToken := req.HeaderParameter(http.CanonicalHeaderKey(model.XPluginJwtHeader))
	if pluginToken == "" {
		return "", nil
	}
	payload, ok, err := s.secretProvider.Parse(strings.TrimPrefix(pluginToken, "Bearer "))
	if err != nil {
		return "", fmt.Errorf("error parse plugin token(%s): %w", pluginToken, err)
	}
	if !ok {
		return "", fmt.Errorf("plugin invalid token(%s)", pluginToken)
	}
	pluginIDInterface, ok := payload["plugin_id"]
	if !ok {
		return "", fmt.Errorf("error plugin token(%s) not found plugin_id", pluginToken)
	}
	pluginID, ok := pluginIDInterface.(string)
	if !ok {
		return "", fmt.Errorf("error plugin token(%s) payload plugin_id(%v) type invalid",
			pluginToken, pluginIDInterface)
	}
	return pluginID, nil
}

func (s *KeelServiceV1) externalGetUser(req *restful.Request) (*model.User, error) {
	token := req.HeaderParameter(http.CanonicalHeaderKey(AuthorizationHeader))
	if token == "" {
		return nil, errors.New("invalid token")
	}
	isManager, err := s.isManagerToken(token)
	if err != nil {
		return nil, fmt.Errorf("error manager token(%s) check: %w", token, err)
	}
	user := new(model.User)
	if !isManager {
		// tKeel platform.
		tKeelToken, err := oauth.GetOauthOperator().ValidationBearerToken(req.Request)
		if err != nil {
			return nil, fmt.Errorf("error validation bearer token(%v): %w",
				req.HeaderParameter(http.CanonicalHeaderKey(AuthorizationHeader)), err)
		}
		tenant := strings.Split(tKeelToken.GetUserID(), "-")[1]
		user.User = tKeelToken.GetUserID()
		user.Tenant = tenant
		// TODO: RBAC.
		user.Role = model.AdminRole
	} else {
		// manager platform.
		user.User = model.TKeelUser
		user.Tenant = model.TKeelTenant
		user.Role = model.AdminRole
	}

	return user, nil
}

func (s *KeelServiceV1) isManagerToken(token string) (bool, error) {
	payload, err := s.secretProvider.ParseUnverified(strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		return false, fmt.Errorf("error parse token(%s): %w", token, err)
	}
	iss, ok := payload["iss"]
	if !ok {
		return false, nil
	}
	if iss == "rudder" {
		_, valid, err := s.secretProvider.Parse(strings.TrimPrefix(token, "Bearer "))
		if err != nil {
			return true, fmt.Errorf("error parse token(%s): %w", token, err)
		}
		if !valid {
			return true, fmt.Errorf("error token(%s) is invalid", token)
		}
		return true, nil
	}
	return false, nil
}

func (s *KeelServiceV1) ProxyAddons(
	resp http.ResponseWriter, req *http.Request) error {
	log.Debugf("proxy call addons %s", req)
	up, err := s.getAddonsUpstream(req)
	if err != nil {
		if errors.Is(err, ErrNotFoundUpstream) {
			writeResult(resp, http.StatusNotFound, "method not found", nil)
		} else {
			writeResult(resp, http.StatusBadRequest, "invail addons", nil)
		}
		return fmt.Errorf("error get addons upstream: %w", err)
	}
	if err = up.Verify(req); err != nil {
		if errors.Is(err, ErrNotActiveUpstream) {
			writeResult(resp, http.StatusUnauthorized, "not active", nil)
		} else {
			writeResult(resp, http.StatusInternalServerError, "internal error", nil)
		}
		return fmt.Errorf("error upstream verify validity: %w", err)
	}
	bodyByte := make([]byte, 0)
	if req.ContentLength != 0 {
		b, err1 := io.ReadAll(req.Body)
		if err1 != nil {
			writeResult(resp, http.StatusBadRequest, err1.Error(), nil)
			return fmt.Errorf("error read body: %w", err1)
		}
		bodyByte = b
		defer req.Body.Close()
	}

	dstResp, err := s.httpDaprClient.Call(req.Context(), &t_dapr.AppRequest{
		ID:         up.ID,
		Method:     up.Method,
		Verb:       req.Method,
		Header:     req.Header,
		QueryValue: req.URL.Query(),
		Body:       bodyByte,
	})
	if err != nil {
		writeResult(resp, http.StatusBadRequest, err.Error(), nil)
		return fmt.Errorf("error plugin client call: %w", err)
	}
	if err = proxyHTTPResponse2RestfulResponse(dstResp, resp); err != nil {
		return fmt.Errorf("error proxy http response 2 restful response: %w", err)
	}
	return nil
}

func (s *KeelServiceV1) ProxyPlugin(
	resp http.ResponseWriter, req *http.Request) error {
	up, err := s.getPluginUpstream(req)
	if err != nil {
		if errors.Is(err, ErrNotFoundUpstream) {
			writeResult(resp, http.StatusNotFound, "upstream not found", nil)
		} else {
			writeResult(resp, http.StatusBadRequest, "invail upstream", nil)
		}
		return fmt.Errorf("error get plugin upstream: %w", err)
	}
	log.Debugf("proxy call plugin %s", up)
	if err = up.Verify(req); err != nil {
		if errors.Is(err, ErrNotActiveUpstream) {
			writeResult(resp, http.StatusUnauthorized, "not active", nil)
		} else {
			writeResult(resp, http.StatusInternalServerError, "internal error", nil)
		}
		return fmt.Errorf("error upstream verify validity: %w", err)
	}
	bodyByte := make([]byte, 0)
	if req.ContentLength != 0 {
		b, err1 := io.ReadAll(req.Body)
		if err1 != nil {
			writeResult(resp, http.StatusBadRequest, err1.Error(), nil)
			return fmt.Errorf("error read body: %w", err1)
		}
		bodyByte = b
		defer req.Body.Close()
	}

	dstResp, err := s.httpDaprClient.Call(req.Context(), &t_dapr.AppRequest{
		ID:         up.ID,
		Method:     up.Method,
		Verb:       req.Method,
		Header:     req.Header,
		QueryValue: req.URL.Query(),
		Body:       bodyByte,
	})
	if err != nil {
		writeResult(resp, http.StatusBadRequest, err.Error(), nil)
		return fmt.Errorf("error plugin client call: %w", err)
	}
	if err = proxyHTTPResponse2RestfulResponse(dstResp, resp); err != nil {
		return fmt.Errorf("error proxy http response 2 restful response: %w", err)
	}
	return nil
}

func (s *KeelServiceV1) ProxyCore(resp http.ResponseWriter, req *http.Request) error {
	log.Debugf("proxy call core %s", req.RequestURI)
	dstPath := strings.TrimPrefix(req.URL.Path, v1.ApisRootPath+v1.CoreSubPath)
	if err := proxyHTTP(req.Context(), s.conf.CoreAddr, dstPath, resp, req); err != nil {
		return fmt.Errorf("error proxy core: %w", err)
	}
	return nil
}

func (s *KeelServiceV1) ProxySecurity(resp http.ResponseWriter, req *http.Request) error {
	log.Debugf("proxy call security %s", req.RequestURI)
	dstPath := strings.TrimPrefix(req.URL.Path, v1.ApisRootPath+v1.SecuritySubPath)
	if err := proxyHTTP(req.Context(), s.conf.RudderAddr, dstPath, resp, req); err != nil {
		return fmt.Errorf("error proxy security: %w", err)
	}
	return nil
}

func (s *KeelServiceV1) ProxyRudder(resp http.ResponseWriter, req *http.Request) error {
	log.Debugf("proxy call rudder %s", req.RequestURI)
	inWhiteList := false
	for _, v := range s.pathWhiteList[v1.RudderSubPath] {
		if v == req.URL.Path {
			inWhiteList = true
			break
		}
	}
	if !inWhiteList {
		user, ok := getUser(req.Context())
		if !ok {
			writeResult(resp, http.StatusUnauthorized, "invalid user", nil)
			return errors.New("error invalid user")
		}
		if user.User != "_tKeel" {
			if req.URL.Path != "/apis/rudder/v1/entries" && user.Role != model.AdminRole {
				writeResult(resp, http.StatusUnauthorized, "invalid role", nil)
				return errors.New("error invalid role")
			}
		}
	}
	dstPath := strings.TrimPrefix(req.URL.Path, v1.ApisRootPath+v1.RudderSubPath)
	if err := proxyHTTP(req.Context(), s.conf.RudderAddr, dstPath, resp, req); err != nil {
		return fmt.Errorf("error proxy rudder: %w", err)
	}
	return nil
}

func (s *KeelServiceV1) getPluginUpstream(req *http.Request) (*upstream, error) {
	pluginID := getPluginIDFromApisPath(req.URL.Path)
	pluginMethod := getMethodApisPath(req.URL.Path)
	if pluginID == "" {
		return nil, fmt.Errorf("get(%s) invalid plugin id", req.URL.Path)
	}
	log.Debugf("pluginID: %s,pluginMethod: %s", pluginID, pluginMethod)
	upstreamRouteInterface, ok := s.pluginRouteMap.Load(pluginID)
	if !ok {
		return nil, ErrNotFoundUpstream
	}
	upstreamRoute, ok := upstreamRouteInterface.(*model.PluginRoute)
	if !ok {
		return nil, errors.New("invalid plugin route type")
	}
	return &upstream{
		ID:            pluginID,
		Method:        pluginMethod,
		TKeelDepened:  upstreamRoute.TkeelVersion,
		ActiveTenants: upstreamRoute.ActiveTenantes,
	}, nil
}

func (s *KeelServiceV1) getAddonsUpstream(req *http.Request) (*upstream, error) {
	addonsMethod := getMethodApisPath(req.URL.Path)
	if addonsMethod == "" {
		return nil, errors.New("invalid addons method")
	}
	src, ok := getSource(req.Context())
	if !ok {
		return nil, errors.New("invalid source")
	}
	srcRouteInterface, ok := s.pluginRouteMap.Load(src.ID)
	if !ok {
		return nil, errors.New("not found source plugin")
	}
	srcRoute, ok := srcRouteInterface.(*model.PluginRoute)
	if !ok {
		return nil, errors.New("invalid source plugin route type")
	}
	upstreamStr, ok := srcRoute.RegisterAddons[addonsMethod]
	if !ok {
		return nil, ErrNotFoundAddons
	}
	upID, upMethod := util.DecodePluginRoute(upstreamStr)
	upRouteInterface, ok := s.pluginRouteMap.Load(upID)
	if !ok {
		return nil, ErrNotFoundUpstream
	}
	upRoute, ok := upRouteInterface.(*model.PluginRoute)
	if !ok {
		return nil, errors.New("invalid upstream plugin route type")
	}
	return &upstream{
		ID:            upID,
		Method:        upMethod,
		ActiveTenants: upRoute.ActiveTenantes,
		TKeelDepened:  upRoute.TkeelVersion,
	}, nil
}

func proxyHTTP(ctx context.Context, host, dstPath string,
	resp http.ResponseWriter, req *http.Request) error {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		writeResult(resp, http.StatusInternalServerError, err.Error(), nil)
		return fmt.Errorf("error read request body: %w", err)
	}
	url := fmt.Sprintf("http://%s%s", host, dstPath)
	if req.URL.RawQuery != "" {
		url += "?" + req.URL.RawQuery
	}
	proxyReq, err := http.NewRequestWithContext(req.Context(), req.Method, url, bytes.NewReader(body))
	if err != nil {
		writeResult(resp, http.StatusInternalServerError, err.Error(), nil)
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
		writeResult(resp, http.StatusBadRequest, err.Error(), nil)
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
		writeResult(resp, http.StatusBadRequest, err.Error(), nil)
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

func getPluginIDFromApisPath(pluginPath string) string {
	ss := strings.SplitN(pluginPath, "/", 4)
	if len(ss) != 4 {
		return ""
	}
	return ss[2]
}

func getMethodApisPath(apisPath string) string {
	ss := strings.SplitN(apisPath, "/", 4)
	if len(ss) != 4 {
		return ""
	}
	return strings.Split(ss[3], "?")[0]
}

func withUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, contextUserKey, user)
}

func getUser(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(contextUserKey).(*model.User)
	if !ok {
		return nil, false
	}
	return user, true
}

func withSource(ctx context.Context, source *source) context.Context {
	return context.WithValue(ctx, contextSourceKey, source)
}

func getSource(ctx context.Context) (*source, bool) {
	src, ok := ctx.Value(contextSourceKey).(*source)
	if !ok {
		return nil, false
	}
	return src, true
}
