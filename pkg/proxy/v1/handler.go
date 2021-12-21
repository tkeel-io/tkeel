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
	"io"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/tkeel-io/kit/log"
	v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/proxy"
	"github.com/tkeel-io/tkeel/pkg/util"
)

// registerContainerHandler register proxy handler.
func registerContainerHandler(c *restful.Container, srv proxy.PluginProxyServer) {
	c.HandleWithFilter(ApisRootPath+CoreSubPath+"/", proxyCore(srv))
	c.HandleWithFilter(ApisRootPath+SecuritySubPath+"/", proxySecurity(srv))
	c.HandleWithFilter(ApisRootPath+RudderSubPath+"/", proxyRudder(srv))
	c.HandleWithFilter(ApisRootPath+"/", proxyPlugin(srv))
	// addons.
	ws := new(restful.WebService)
	ws.Path(AddonsRootPath)
	registerProxyPath(ws, "/{"+AddonsNamePath+"}", proxyAddons(srv))
}

// proxyCore call core.
func proxyCore(srv proxy.PluginProxyServer) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		log.Debug("proxy core")
		if err := srv.ProxyCore(req.Context(), rw, req); err != nil {
			log.Errorf("error proxy core err: %s", err)
		}
	}
}

// proxyRudder call rudder.
func proxyRudder(srv proxy.PluginProxyServer) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		log.Debug("proxy rudder")
		if err := srv.ProxyRudder(req.Context(), rw, req); err != nil {
			log.Errorf("error proxy rudder err: %s", err)
		}
	}
}

// proxySecurity call Security.
func proxySecurity(srv proxy.PluginProxyServer) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		log.Debug("proxy Security")
		if err := srv.ProxySecurity(req.Context(), rw, req); err != nil {
			log.Errorf("error proxy Security err: %s", err)
		}
	}
}

// proxyPlugin call the request to the corresponding plugin method.
func proxyPlugin(srv proxy.PluginProxyServer) http.HandlerFunc {
	return func(responseWrite http.ResponseWriter, req *http.Request) {
		sub := getSubPath(req.URL.Path, ApisRootPath)
		defer req.Body.Close()
		b, err := io.ReadAll(req.Body)
		if err != nil {
			log.Errorf("error read request body: %s", err)
			responseWrite.WriteHeader(http.StatusBadRequest)
			responseWrite.Write([]byte(err.Error()))
			return
		}
		var pRoute *model.PluginRoute
		pluginID := getPluginIDFromApisPath(sub)
		method := getPluginMethodApisPath(sub)
		pRouteIn, ok := pluginRouteMap.Load(pluginID)
		if !ok {
			log.Debugf("plugin(%s) not register", pluginID)
			responseWrite.WriteHeader(http.StatusNotFound)
			responseWrite.Write([]byte("not registered"))
			return
		}
		pRoute, ok = pRouteIn.(*model.PluginRoute)
		if !ok {
			log.Errorf("error plugin(%s) route type not model.PluginRoute point", pluginID)
			responseWrite.WriteHeader(http.StatusInternalServerError)
			responseWrite.Write([]byte("internal error"))
			return
		}
		if pRoute.Status != v1.PluginStatus_RUNNING {
			if pRoute.Status != v1.PluginStatus_UNREGISTER {
				log.Errorf("error plugin(%s) status(%s) not running", pluginID, pRoute.Status)
				responseWrite.WriteHeader(http.StatusForbidden)
				responseWrite.Write([]byte("plugin cannot provide services"))
				return
			}
			srcIDIn := req.Context().Value(ContextPluginIDKey)
			srcID, ok := srcIDIn.(string)
			if !ok {
				log.Errorf("error plugin(%s) status(%s) not running", pluginID, pRoute.Status)
				responseWrite.WriteHeader(http.StatusForbidden)
				responseWrite.Write([]byte("plugin cannot provide services"))
				return
			}
			in := false
			for _, v := range pRoute.ImplementedPlugin {
				if v == srcID {
					in = true
				}
			}
			if !in {
				log.Errorf("error plugin(%s) status(%s) not running", pluginID, pRoute.Status)
				responseWrite.WriteHeader(http.StatusForbidden)
				responseWrite.Write([]byte("plugin cannot provide services"))
				return
			}
		}
		proxyReq := &proxy.Reqeust{
			ID:         pluginID,
			Method:     method,
			Verb:       req.Method,
			QueryValue: req.URL.Query(),
			Header:     req.Header,
			Body:       b,
		}
		if err = srv.ProxyPlugin(req.Context(), responseWrite, proxyReq); err != nil {
			log.Errorf("error proxy plugin: %s", err)
		}
	}
}

// proxyAddons call the request to the corresponding plugin addons.
func proxyAddons(srv proxy.PluginProxyServer) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		addons := req.PathParameter(AddonsNamePath)
		srcPluginIDIn := req.Attribute(SrcPluginIDAttribute)
		srcPluginID, ok := srcPluginIDIn.(string)
		if !ok {
			log.Errorf("error invaild type src plugin id: %v", srcPluginIDIn)
			resp.WriteErrorString(http.StatusInternalServerError,
				"src plugin type error")
			return
		}
		// check src.
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
		if pr.Status != v1.PluginStatus_RUNNING {
			log.Errorf("error src plugin(%s) status(%s) not running", srcPluginID, pr.Status)
			resp.WriteErrorString(http.StatusForbidden,
				"plugin cannot provide services")
			return
		}
		// get dest path.
		dstPath, ok := pr.RegisterAddons[addons]
		if !ok {
			log.Debugf("plugin(%s) addons(%s) not Implemented",
				srcPluginID, addons)
			resp.WriteErrorString(http.StatusNotFound,
				"Addons not implemented")
			return
		}
		dstPluginID, dstMethod := util.DecodePluginRoute(dstPath)
		// check dest.
		dPrI, ok := pluginRouteMap.Load(dstPluginID)
		if !ok {
			log.Errorf("error not found dst plugin(%s) route", dstPluginID)
			resp.WriteErrorString(http.StatusInternalServerError,
				"not found src plugin("+srcPluginID+")")
			return
		}
		dPr, ok := dPrI.(*model.PluginRoute)
		if !ok {
			log.Errorf("error dst pri not plugin route")
			resp.WriteErrorString(http.StatusInternalServerError,
				"internal error")
			return
		}
		if dPr.Status != v1.PluginStatus_RUNNING {
			log.Errorf("error dst plugin(%s) status(%s) not running", dstPluginID, pr.Status)
			resp.WriteErrorString(http.StatusForbidden,
				"plugin cannot provide services")
			return
		}
		// proxy.
		b, err := io.ReadAll(req.Request.Body)
		if err != nil {
			log.Errorf("error read request body: %s", err)
			if err1 := resp.WriteError(http.StatusBadRequest, err); err1 != nil {
				log.Errorf("error write error(%d,%s): %s", http.StatusBadRequest, err1)
			}
			return
		}
		proxyReq := &proxy.Reqeust{
			ID:         dstPluginID,
			Method:     dstMethod,
			Verb:       req.Request.Method,
			QueryValue: req.Request.URL.Query(),
			Header:     req.Request.Header,
			Body:       b,
		}
		if err = srv.ProxyPlugin(req.Request.Context(), resp.ResponseWriter, proxyReq); err != nil {
			log.Errorf("error proxy plugin: %s", err)
		}
	}
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
