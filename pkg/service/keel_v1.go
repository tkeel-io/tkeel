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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/tkeel-io/tkeel/pkg/client"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	v1 "github.com/tkeel-io/tkeel/pkg/service/keel/v1"
	"github.com/tkeel-io/tkeel/pkg/token"
	"github.com/tkeel-io/tkeel/pkg/util"
	"github.com/tkeel-io/tkeel/pkg/version"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/emicklei/go-restful"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/kit/result"

	oauth_pb "github.com/tkeel-io/tkeel/api/security_oauth/v1"
	tenant_pb "github.com/tkeel-io/tkeel/api/tenant/v1"
)

const (
	AuthorizationHeader = "Authorization"

	_securityComponent               = "rudder"
	_securityAuthenticate            = "/v1/oauth/authenticate"
	_securityTenantPluginPermissible = "/v1/tenants/plugins/permissible"
)

var (
	contextSessionKey    = struct{}{}
	ErrNotFoundUpstream  = errors.New("not found upstream plugin")
	ErrNotActiveUpstream = errors.New("not active upstream plugin")
	ErrNotFoundAddons    = errors.New("not found addons")
)

type session struct {
	Src           *endpoint
	Dst           *endpoint
	User          *model.User
	RequestMethod string
}

func (s *session) String() string {
	b, err := json.Marshal(s)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

type endpoint struct {
	ID           string
	TKeelDepened string
}

func (e *endpoint) String() string {
	b, err := json.Marshal(e)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

type KeelServiceV1 struct {
	watchInterval   string
	conf            *config.ProxyConf
	httpDaprClient  dapr.Client
	pluginRouteOp   proute.Operator
	pluginRouteMap  *sync.Map
	secretProvider  token.Provider
	timeout         time.Duration
	regExpWhiteList []string
}

func NewKeelServiceV1(interval string, conf *config.Configuration, client dapr.Client, op proute.Operator) *KeelServiceV1 {
	secret := util.RandStringBytesMaskImpr(0, 10)
	duration, err := time.ParseDuration(conf.Proxy.Timeout)
	if err != nil {
		log.Errorf("error parse duration(%s): %s", conf.Proxy.Timeout, err)
		duration = 10 * time.Second
	}
	ksV1 := &KeelServiceV1{
		watchInterval:  interval,
		conf:           conf.Proxy,
		httpDaprClient: client,
		pluginRouteOp:  op,
		pluginRouteMap: new(sync.Map),
		secretProvider: token.InitProvider(secret, "", ""),
		timeout:        duration,
		regExpWhiteList: []string{
			"/apis/rudder/v1/oauth2*",
			"/apis/security/v1/oauth*",
		},
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
		sess := new(session)
		// dst and method.
		if err := s.setDstAndMethod(ctx, sess, resp, req); err != nil {
			log.Errorf("error set dst and method session: %s", err)
			return
		}
		// white path.
		if !s.matchRegExpWhiteList(req.Request.URL.Path) {
			// src and user session.
			if err := s.setSrcAndUserSession(ctx, sess, resp, req.Request.Header); err != nil {
				log.Errorf("error set source and user session: %s", err)
				return
			}
			// check sess.
			if err := s.checkSession(ctx, sess, resp); err != nil {
				log.Errorf("error check session: %s", err)
				return
			}
			// set header.
			req.Request.Header[model.XtKeelAuthHeader] = []string{sess.User.Base64Encode()}
		}
		// set context
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

func (s *KeelServiceV1) setSrcAndUserSession(ctx context.Context, sess *session, resp *restful.Response, header http.Header) error {
	// set src plugin id.
	pluginID, err := s.checkXPluginJwt(ctx, header)
	if err != nil {
		writeResult(resp, http.StatusUnauthorized, "invalid x-plugin-jwt token")
		return fmt.Errorf("error get plugin ID from request: %w", err)
	}
	sess.Src = &endpoint{
		ID: pluginID,
	}
	// set user.
	if pluginID == "" {
		log.Debugf("external flow")
		sess.User, err = s.checkAuthorization(ctx, header)
		if err != nil {
			writeResult(resp, http.StatusUnauthorized, "invalid authorization token")
			return fmt.Errorf("error external get user: %w", err)
		}
		sess.Src.TKeelDepened = version.Version
		return nil
	}

	log.Debugf("internal flow")
	sess.User, err = s.checkXtKeelAtuh(ctx, header)
	if err != nil {
		writeResult(resp, http.StatusUnauthorized, "invalid x-tKeel-auth token")
		return fmt.Errorf("error internal get user: %w", err)
	}
	if pluginID != "rudder" &&
		pluginID != "core" &&
		pluginID != "security" &&
		pluginID != "keel" {
		pluginRouteInterface, ok := s.pluginRouteMap.Load(pluginID)
		if !ok {
			writeResult(resp, http.StatusUnauthorized, "invalid plugin")
			return fmt.Errorf("error source plugin ID(%s) not register", pluginID)
		}
		pluginRoute, ok := pluginRouteInterface.(*model.PluginRoute)
		if !ok {
			writeResult(resp, http.StatusInternalServerError, "internal error")
			return errors.New("error source plugin route type invalid")
		}
		sess.Src.TKeelDepened = pluginRoute.TkeelVersion
	} else {
		sess.Src.TKeelDepened = version.Version
	}
	return nil
}

func (s *KeelServiceV1) setDstAndMethod(ctx context.Context, sess *session, resp *restful.Response, req *restful.Request) error {
	sess.Dst = &endpoint{}
	if isAddons(req.Request.URL.Path) {
		if err := s.getAddonsDstAndMethod(sess, req.Request); err != nil {
			if errors.Is(err, ErrNotFoundAddons) || errors.Is(err, ErrNotActiveUpstream) {
				writeResult(resp, http.StatusNotFound, "addons not found")
			}
			writeResult(resp, http.StatusBadGateway, "error proxy addons")
			return fmt.Errorf("error get(%s) addons destination and method", sess)
		}
	} else {
		if err := s.getPluginDstAndMethod(sess, req.Request); err != nil {
			writeResult(resp, http.StatusBadGateway, "error proxy plugin")
			return fmt.Errorf("error get(%s) plugin destination and method", sess)
		}
	}
	return nil
}

func (s *KeelServiceV1) checkSession(ctx context.Context, sess *session, resp *restful.Response) error {
	res, err := s.callTenantPluginPermissible(ctx, sess)
	if err != nil {
		writeResult(resp, http.StatusBadGateway, "error check plugin tenant")
		return fmt.Errorf("error call tenant plugin permissible: %w", err)
	}
	callResp := &tenant_pb.PluginPermissibleResponse{}
	if err = anypb.UnmarshalTo(res.Data, callResp, proto.UnmarshalOptions{}); err != nil {
		writeResult(resp, http.StatusBadGateway, "error check plugin tenant")
		return fmt.Errorf("error unmarshal authorization response(%s) check: %w", res, err)
	}
	if !callResp.Allowed {
		writeResult(resp, http.StatusUnauthorized, "upstream plugin not enable")
		return fmt.Errorf("error upstream plugin(%s/%s) not enable", sess.Dst.ID, sess.User.Tenant)
	}
	return nil
}

func (s *KeelServiceV1) checkXPluginJwt(ctx context.Context, header http.Header) (string, error) {
	// TODO: webhook.
	pluginToken := header.Get(model.XPluginJwtHeader)
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

func (s *KeelServiceV1) checkAuthorization(ctx context.Context, header http.Header) (*model.User, error) {
	authorization := header.Get(model.AuthorizationHeader)
	if authorization == "" {
		return nil, errors.New("invalid authorization")
	}
	isManager, err := s.isManagerToken(authorization)
	if err != nil {
		return nil, fmt.Errorf("error manager authorization(%s) check: %w", authorization, err)
	}
	user := new(model.User)
	if !isManager {
		// TODO payload.
		h := make(http.Header)
		h.Set(model.AuthorizationHeader, authorization)
		res, err1 := s.callAuthorization(ctx, _securityComponent, _securityAuthenticate, h)
		if err1 != nil {
			return nil, fmt.Errorf("error call Authorization(%s/%s) check: %w",
				_securityComponent, _securityAuthenticate, err1)
		}
		resp := &oauth_pb.AuthenticateResponse{}
		if err = anypb.UnmarshalTo(res.Data, resp, proto.UnmarshalOptions{}); err != nil {
			return nil, fmt.Errorf("error unmarshal authorization response(%s) check: %w", res, err)
		}
		user.User = resp.GetUserId()
		user.Tenant = resp.GetTenantId()
		user.Role = model.AdminRole
	} else {
		if err = s.verifyManagerToken(authorization); err != nil {
			return nil, fmt.Errorf("error verify manager token(%s): %w", authorization, err)
		}
		user.User = model.TKeelUser
		user.Tenant = model.TKeelTenant
		user.Role = model.AdminRole
	}
	return user, nil
}

func (s *KeelServiceV1) checkXtKeelAtuh(ctx context.Context, header http.Header) (*model.User, error) {
	xtKeelAuthString := header.Get(model.XtKeelAuthHeader)
	if xtKeelAuthString == "" {
		return nil, errors.New("invalid x-tKeel-auth")
	}
	user := new(model.User)
	if err := user.Base64Decode(xtKeelAuthString); err != nil {
		return nil, fmt.Errorf("error decode x-tKeel-auth(%s): %w", xtKeelAuthString, err)
	}
	return user, nil
}

func (s *KeelServiceV1) callAuthorization(ctx context.Context, componentID, method string, header http.Header) (*result.Http, error) {
	log.Debugf("call authenticate %s/%s/%s", componentID, method, header)
	out, err := client.InvokeJSON(ctx, s.httpDaprClient, &dapr.AppRequest{
		ID:         componentID,
		Method:     method,
		Verb:       http.MethodGet,
		Header:     header,
		QueryValue: nil,
		Body:       nil,
	}, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error invoke json: %w", err)
	}
	res := &result.Http{}
	if err = protojson.Unmarshal(out, res); err != nil {
		return nil, fmt.Errorf("error protojson unmarshal: %w", err)
	}
	if res.Code != http.StatusOK {
		return nil, fmt.Errorf("error result: %s", res)
	}
	return res, nil
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
		return true, nil
	}
	return false, nil
}

func (s *KeelServiceV1) verifyManagerToken(token string) error {
	_, valid, err := s.secretProvider.Parse(strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		return fmt.Errorf("error parse token(%s): %w", token, err)
	}
	if !valid {
		return fmt.Errorf("error token(%s) is invalid", token)
	}
	return nil
}

func (s *KeelServiceV1) matchRegExpWhiteList(path string) bool {
	for _, v := range s.regExpWhiteList {
		match, err := regexp.MatchString(v, path)
		if err != nil {
			log.Errorf("error regular expression(%s/%s) run: %s", v, path, err)
			return false
		}
		if match {
			return true
		}
	}
	return false
}

func (s *KeelServiceV1) getAddonsDstAndMethod(sess *session, req *http.Request) error {
	addonsMethod := getMethodApisPath(req.URL.Path)
	if addonsMethod == "" {
		return errors.New("invalid addons method")
	}
	srcRouteInterface, ok := s.pluginRouteMap.Load(sess.Src.ID)
	if !ok {
		return errors.New("not found source plugin")
	}
	srcRoute, ok := srcRouteInterface.(*model.PluginRoute)
	if !ok {
		return errors.New("invalid source plugin route type")
	}
	dstStr, ok := srcRoute.RegisterAddons[addonsMethod]
	if !ok {
		return ErrNotFoundAddons
	}
	dstID, dstMethod := util.DecodePluginRoute(dstStr)
	upRouteInterface, ok := s.pluginRouteMap.Load(dstID)
	if !ok {
		return ErrNotFoundUpstream
	}
	dstRoute, ok := upRouteInterface.(*model.PluginRoute)
	if !ok {
		return errors.New("invalid upstream plugin route type")
	}
	sess.Dst.ID = dstID
	sess.RequestMethod = dstMethod
	sess.Dst.TKeelDepened = dstRoute.TkeelVersion
	return nil
}

func (s *KeelServiceV1) getPluginDstAndMethod(sess *session, req *http.Request) error {
	pluginID := getPluginIDFromApisPath(req.URL.Path)
	pluginMethod := getMethodApisPath(req.URL.Path)
	if pluginID == "" {
		return fmt.Errorf("get(%s) invalid plugin id", req.URL.Path)
	}
	log.Debugf("pluginID: %s,pluginMethod: %s", pluginID, pluginMethod)
	if pluginID == "core" ||
		pluginID == "keel" ||
		pluginID == "rudder" ||
		pluginID == "security" {
		// convert security ==> rudder.
		if pluginID == "security" {
			pluginID = "rudder"
		}
		sess.Dst.TKeelDepened = version.Version
	} else {
		upstreamRouteInterface, ok := s.pluginRouteMap.Load(pluginID)
		if !ok {
			return ErrNotFoundUpstream
		}
		upstreamRoute, ok := upstreamRouteInterface.(*model.PluginRoute)
		if !ok {
			return errors.New("invalid plugin route type")
		}
		sess.Dst.TKeelDepened = upstreamRoute.TkeelVersion
	}
	sess.Dst.ID = pluginID
	sess.RequestMethod = pluginMethod
	return nil
}

func (s *KeelServiceV1) callTenantPluginPermissible(ctx context.Context, sess *session) (*result.Http, error) {
	log.Debugf("call TenantPluginPermissible %s/%s", sess.Dst.ID, sess.User.Tenant)
	v := make(url.Values)
	v.Add("tenant_id", sess.User.Tenant)
	v.Add("plugin_id", sess.Dst.ID)
	out, err := client.InvokeJSON(ctx, s.httpDaprClient, &dapr.AppRequest{
		ID:         _securityComponent,
		Method:     _securityTenantPluginPermissible,
		Verb:       http.MethodGet,
		Header:     make(http.Header),
		QueryValue: v,
		Body:       nil,
	}, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error invoke json: %w", err)
	}
	res := &result.Http{}
	if err = protojson.Unmarshal(out, res); err != nil {
		return nil, fmt.Errorf("error protojson unmarshal: %w", err)
	}
	if res.Code != http.StatusOK {
		return nil, fmt.Errorf("error result: %s", res)
	}
	return res, nil
}

func isAddons(path string) bool {
	return strings.HasPrefix(path, v1.ApisRootPath+v1.AddonsRootPath)
}

func withSession(ctx context.Context, sess *session) context.Context {
	return context.WithValue(ctx, contextSessionKey, sess)
}

func getSession(ctx context.Context) (*session, bool) {
	src, ok := ctx.Value(contextSessionKey).(*session)
	if !ok {
		return nil, false
	}
	return src, true
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
