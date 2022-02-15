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
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/service/keel"
)

// registerContainerHandler register proxy handler.
func registerContainerHandler(c *restful.Container, srv keel.ProxyServer) {
	c.HandleWithFilter(ApisRootPath+"/", proxyPlugin(srv))
	c.HandleWithFilter(StaticRootPath+"/", proxyPlugin(srv))
}

// proxyPlugin call the request to the corresponding plugin method.
func proxyPlugin(srv keel.ProxyServer) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		log.Debug("proxy Plugin")
		if err := srv.ProxyPlugin(rw, req); err != nil {
			log.Errorf("error proxy plugin: %s", err)
		}
	}
}
