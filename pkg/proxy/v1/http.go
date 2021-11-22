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

package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/emicklei/go-restful"
	"github.com/golang-jwt/jwt"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/util"
)

const (
	ApisRootPath   = "/apis"
	XKeelHeader    = "x-plugin-jwt"
	MethodPath     = "method"
	AddonsNamePath = "addons_name"
)

var pluginRouteMap = new(sync.Map)

func RegisterPluginProxyHTTPServer(ctx context.Context, container *restful.Container,
	srv PluginProxyServer) error {
	if container == nil {
		return errors.New("error invaild container: nil")
	}
	// container.Filter().
	go srv.Watch(ctx, func(pprm model.PluginProxyRouteMap) error {
		wsList := container.RegisteredWebServices()
		var apisWs *restful.WebService
		for _, ws := range wsList {
			if ws.RootPath() == ApisRootPath {
				apisWs = ws
				break
			}
		}
		if apisWs == nil {
			apisWs = new(restful.WebService)
			apisWs.Filter(proxyPluginFilter)
			container.Add(apisWs)
		}
		if err := updateWebServiceRoute(apisWs, srv, pprm); err != nil {
			return fmt.Errorf("error update web service route: %w", err)
		}
		return nil
	})
	return nil
}

func proxyPluginFilter(req *restful.Request, resp *restful.Response,
	chain *restful.FilterChain) {
	var srcPluginID, dstPluginID, method string
	token := req.HeaderParameter(XKeelHeader)
	log.Debugf("token: %s", token)
	if token != "" {
		s, err := checkToken(token)
		if err != nil {
			log.Errorf("error check token(%s): %s", token, err)
			if err = resp.WriteErrorString(http.StatusForbidden,
				"x-plugin-jwt token invaild"); err != nil {
				log.Errorf("error response write error string(%d,%s): %s",
					http.StatusForbidden, "x-plugin-jwt token invaild", err)
			}
			return
		}
		srcPluginID = s
	}
	sub := getSubPath(req.Request.URL.Path, ApisRootPath)
	dstPluginID = convertPluginPath2PluginID(sub)
	method = req.PathParameter(MethodPath)

	if srcPluginID != "" {
		srcPrI, ok := pluginRouteMap.Load(srcPluginID)
		if !ok {
			log.Errorf("error not found src plugin(%s) route", srcPluginID)
			resp.WriteErrorString(http.StatusInternalServerError,
				"not found src plugin("+srcPluginID+")")
			return
		}
		srcPr, ok := srcPrI.(*model.PluginRoute)
		if !ok {
			log.Errorf("error src pri not plugin route")
			resp.WriteErrorString(http.StatusInternalServerError,
				"internal error")
			return
		}

		dstPrI, ok := pluginRouteMap.Load(dstPluginID)
		if !ok {
			log.Errorf("error not found dst plugin(%s) route", dstPluginID)
			resp.WriteErrorString(http.StatusInternalServerError,
				"not found dst plugin("+dstPluginID+")")
			return
		}
		dstPr, ok := dstPrI.(*model.PluginRoute)
		if !ok {
			log.Errorf("error dst pri not plugin route")
			resp.WriteErrorString(http.StatusInternalServerError,
				"internal error")
			return
		}

		ok, err := util.CheckRegisterPluginTkeelVersion(dstPr.TkeelVersion,
			srcPr.TkeelVersion)
		if err != nil {
			log.Errorf("error check tkeel version(%s/%s): %s",
				dstPr.TkeelVersion, srcPr.TkeelVersion, err)
			resp.WriteErrorString(http.StatusInternalServerError,
				"internal error")
			return
		}
		if !ok {
			err = fmt.Errorf("error dst plugin tkeel version(%s) not less then src plugin tkeel version(%s)",
				dstPr.TkeelVersion, srcPr.TkeelVersion)
			log.Error(err)
			resp.WriteError(http.StatusForbidden, err)
			return
		}
	}

	log.Debugf("src(%s) request dst(%s) method(%s)",
		srcPluginID, dstPluginID, method)
	chain.ProcessFilter(req, resp)
}

// callPlugin call the request to the corresponding plugin method.
func callPlugin(srv PluginProxyServer) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		sub := getSubPath(req.Request.URL.Path, ApisRootPath)
		defer req.Request.Body.Close()
		b, err := io.ReadAll(req.Request.Body)
		if err != nil {
			log.Errorf("error read request body: %s", err)
			if err1 := resp.WriteError(http.StatusBadRequest, err); err1 != nil {
				log.Errorf("error write error(%d,%s): %s", http.StatusBadRequest, err1)
			}
			return
		}
		proxyReq := &ProxyReqeust{
			ID:         convertPluginPath2PluginID(sub),
			Method:     req.PathParameter("method"),
			QueryValue: req.Request.URL.Query(),
			Header:     req.Request.Header,
			Body:       b,
		}
		callDstPlugin(req.Request.Context(), srv, proxyReq, resp)
	}
}

// callAddons call the request to the corresponding plugin addons.
func callAddons(srv PluginProxyServer) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		sub := getSubPath(req.Request.URL.Path, ApisRootPath)
		srcPluginID := convertPluginPath2PluginID(sub)
		addons := req.PathParameter(AddonsNamePath)
		prI, ok := pluginRouteMap.Load(srcPluginID)
		if !ok {
			log.Errorf("error not found src plugin(%s) route", srcPluginID)
			resp.WriteErrorString(http.StatusInternalServerError,
				"not found src plugin("+srcPluginID+")")
			return
		}
		pr, ok := prI.(*model.PluginRoute)
		if !ok {
			log.Errorf("error src pri not plugin route")
			resp.WriteErrorString(http.StatusInternalServerError,
				"internal error")
			return
		}
		dstPath, ok := pr.RegisterAddons[addons]
		if !ok {
			log.Debugf("plugin(%s) addons(%s) not Implemented",
				srcPluginID, addons)
			resp.WriteErrorString(http.StatusNotFound,
				"Addons not implemented")
			return
		}
		dstPluginID, dstMethod := util.DecodePluginRoute(dstPath)
		b, err := io.ReadAll(req.Request.Body)
		if err != nil {
			log.Errorf("error read request body: %s", err)
			if err1 := resp.WriteError(http.StatusBadRequest, err); err1 != nil {
				log.Errorf("error write error(%d,%s): %s", http.StatusBadRequest, err1)
			}
			return
		}
		proxyReq := &ProxyReqeust{
			ID:         dstPluginID,
			Method:     dstMethod,
			QueryValue: req.Request.URL.Query(),
			Header:     req.Request.Header,
			Body:       b,
		}
		callDstPlugin(req.Request.Context(), srv, proxyReq, resp)
	}
}

func updateWebServiceRoute(ws *restful.WebService, srv PluginProxyServer,
	pprm model.PluginProxyRouteMap) error {
	// add route path and update addons proxy.
	oldRouteMap := routeSlice2Map(ws.Routes())
	for pID, pRoute := range pprm {
		plugnPath := convertPluginID2PluginPath(pID)
		addonsPath := convertPluginID2PluginAddonsPath(pID)
		if _, ok := oldRouteMap[plugnPath]; !ok {
			registerProxyPath(ws, plugnPath, callPlugin(srv))
			registerProxyPath(ws, addonsPath, callAddons(srv))
		}
		pluginRouteMap.Store(pID, pRoute)
	}
	// remove route path.
	newRouteMap := routeSlice2Map(ws.Routes())
	for pluginPath := range newRouteMap {
		pID := convertPluginPath2PluginID(pluginPath)
		if _, ok := pprm[pID]; !ok {
			if err := deletePath(ws, pluginPath); err != nil {
				return fmt.Errorf("error delete plugin path(%s): %w", pluginPath, err)
			}
			addonsPath := convertPluginID2PluginAddonsPath(pID)
			if err := deletePath(ws, addonsPath); err != nil {
				return fmt.Errorf("error delete addons path(%s): %w", addonsPath, err)
			}
			pluginRouteMap.Delete(pID)
		}
	}
	return nil
}

func registerProxyPath(ws *restful.WebService, path string,
	function restful.RouteFunction) {
	ws.Route(ws.GET(path).To(function))
	ws.Route(ws.DELETE(path).To(function))
	ws.Route(ws.POST(path).To(function))
	ws.Route(ws.PUT(path).To(function))
	ws.Route(ws.OPTIONS(path).To(function))
	ws.Route(ws.PATCH(path).To(function))
}

func deletePath(ws *restful.WebService, path string) error {
	var err error
	if err = ws.RemoveRoute(path, http.MethodGet); err != nil {
		return fmt.Errorf("error remove route(%s/%s): %w", path,
			http.MethodGet, err)
	}
	if err = ws.RemoveRoute(path, http.MethodDelete); err != nil {
		return fmt.Errorf("error remove route(%s/%s): %w", path,
			http.MethodDelete, err)
	}
	if err = ws.RemoveRoute(path, http.MethodPost); err != nil {
		return fmt.Errorf("error remove route(%s/%s): %w", path,
			http.MethodPost, err)
	}
	if err = ws.RemoveRoute(path, http.MethodPut); err != nil {
		return fmt.Errorf("error remove route(%s/%s): %w", path,
			http.MethodPut, err)
	}
	if err = ws.RemoveRoute(path, http.MethodOptions); err != nil {
		return fmt.Errorf("error remove route(%s/%s): %w", path,
			http.MethodOptions, err)
	}
	if err = ws.RemoveRoute(path, http.MethodPatch); err != nil {
		return fmt.Errorf("error remove route(%s/%s): %w", path,
			http.MethodPatch, err)
	}
	return nil
}

func getSubPath(path, rootPath string) string {
	return strings.TrimPrefix(path, rootPath)
}

func routeSlice2Map(rs []restful.Route) map[string]restful.Route {
	retMap := make(map[string]restful.Route)
	for _, r := range rs {
		subPath := getSubPath(r.Path, ApisRootPath)
		if _, ok := retMap[subPath]; !ok {
			retMap[subPath] = r
		}
	}
	return retMap
}

func convertPluginID2PluginPath(pluginID string) string {
	return "/" + pluginID + "/{" + MethodPath + "}"
}

func convertPluginID2PluginAddonsPath(pluginID string) string {
	return "/" + pluginID + "/addons/{" + AddonsNamePath + "}"
}

func convertPluginPath2PluginID(pluginPath string) string {
	ss := strings.Split(pluginPath, "/")
	return ss[0]
}

func checkToken(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("error token invaild: %s", token)
	}
	ss := strings.Split(token, " ")
	if len(ss) != 2 {
		return "", fmt.Errorf("error token invaild: %s", token)
	}
	payload := ss[1]
	b, err := jwt.DecodeSegment(payload)
	if err != nil {
		return "", fmt.Errorf("error jwt decode: %w", err)
	}
	pmap := make(map[string]interface{})
	err = json.Unmarshal(b, &pmap)
	if err != nil {
		return "", fmt.Errorf("error json unmarshal: %w", err)
	}
	pID, ok := pmap["plugin_id"]
	if !ok {
		return "", fmt.Errorf("error token(%s) not has field: plugin_id", string(b))
	}
	pIDStr, ok := pID.(string)
	if !ok {
		return "", fmt.Errorf("error token(%s) type invaild", string(b))
	}
	return pIDStr, nil
}

func callDstPlugin(ctx context.Context, srv PluginProxyServer, proxyReq *ProxyReqeust, resp *restful.Response) {
	dstResp, err := srv.Call(ctx, proxyReq)
	if err != nil {
		log.Errorf("error proxy server call(%s): %s",
			proxyReq, err)
		resp.WriteErrorString(http.StatusInternalServerError,
			"internal error")
		return
	}

	for k, vs := range dstResp.Header {
		for _, v := range vs {
			resp.AddHeader(k, v)
		}
	}
	dstBody, err := io.ReadAll(dstResp.Body)
	defer dstResp.Body.Close()
	if err != nil {
		log.Errorf("error read dst response body: %s", err)
		if err1 := resp.WriteError(http.StatusBadRequest, err); err1 != nil {
			log.Errorf("error write error(%d,%s): %s", http.StatusBadRequest, err1)
		}
		return
	}
	resp.WriteHeader(dstResp.StatusCode)
	if len(dstBody) == 0 {
		_, err = resp.Write([]byte(dstResp.Status))
		if err != nil {
			log.Errorf("error write: %s", err)
		}
		return
	}

	go func() {
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
	}()
}
