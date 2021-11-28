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
	"errors"
	"fmt"
	"sync"

	"github.com/emicklei/go-restful"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/proxy"
)

const (
	SecuritySubPath      = "/security"
	RudderSubPath        = "/rudder"
	CoreSubPath          = "/core"
	ApisRootPath         = "/apis"
	AddonsRootPath       = "/addons"
	XKeelHeader          = "x-plugin-jwt"
	MethodPath           = "method"
	AddonsNamePath       = "addons_name"
	SrcPluginIDAttribute = "SrcPluginID"
)

var pluginRouteMap = new(sync.Map)

func RegisterPluginProxyHTTPServer(ctx context.Context, container *restful.Container, conf *config.Configuration,
	externalFilter func(*restful.Request, *restful.Response, *restful.FilterChain), srv proxy.PluginProxyServer) error {
	if container == nil {
		return errors.New("error invaild container: nil")
	}
	whiteList := []string{ApisRootPath + SecuritySubPath, ApisRootPath + RudderSubPath}
	container.Filter(containerFileter(whiteList, externalFilter))
	// register.
	registerContainerHandler(container, srv)
	cb := func(pprm model.PluginProxyRouteMap) error {
		if err := updateWebServiceRoute(srv, pprm); err != nil {
			return fmt.Errorf("error update web service route: %w", err)
		}
		return nil
	}
	// watch plugin route change.
	go srv.Watch(ctx, cb)
	return nil
}

func updateWebServiceRoute(srv proxy.PluginProxyServer,
	pprm model.PluginProxyRouteMap) error {
	log.Debugf("update plugin proxy route map: %s", pprm)
	// upsert new route map.
	for id, v := range pprm {
		pluginRouteMap.Store(id, v)
	}
	// delete old route map.
	pluginRouteMap.Range(func(key, value interface{}) bool {
		pID, ok := key.(string)
		if !ok {
			pluginRouteMap.Delete(key)
			log.Errorf("error invaild key type: %v", key)
			return true
		}
		if _, ok = pprm[pID]; !ok {
			pluginRouteMap.Delete(key)
		}
		return true
	})
	return nil
}
