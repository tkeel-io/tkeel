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

	"github.com/pkg/errors"

	"github.com/emicklei/go-restful"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/service/keel"
)

const (
	ApisRootPath   = "/apis"
	StaticRootPath = "/static"
	AddonsRootPath = "/addons"
)

func RegisterPluginProxyHTTPServer(ctx context.Context,
	container *restful.Container, srv keel.ProxyServer,
) error {
	if container == nil {
		return errors.New("error invalid container: nil")
	}
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{
			restful.HEADER_Accept,
			restful.HEADER_ContentEncoding,
			model.AuthorizationHeader,
			model.XtKeelAuthHeader,
			model.XPluginJwtHeader,
			model.ContentTypeHeader,
		},
		AllowedMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "OPTIONS"},
		CookiesAllowed: true,
		Container:      container,
	}
	container.Filter(cors.Filter)
	container.Filter(container.OPTIONSFilter)
	container.Filter(srv.Filter())
	// register.
	registerContainerHandler(container, srv)
	return nil
}
