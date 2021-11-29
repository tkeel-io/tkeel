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
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/util"
)

func containerFileter(rootPahtWhiteList []string, externalFilter func(*restful.Request,
	*restful.Response, *restful.FilterChain)) func(*restful.Request, *restful.Response, *restful.FilterChain) {
	return func(req *restful.Request, resp *restful.Response,
		chain *restful.FilterChain) {
		// pass wite list.
		for _, v := range rootPahtWhiteList {
			if strings.HasPrefix(req.Request.URL.Path, v) {
				chain.ProcessFilter(req, resp)
				return
			}
		}
		// get src plugin id.
		if proxyPluginTokenFilter(req, resp) {
			// pass addons reqeust.
			if strings.HasPrefix(req.Request.URL.Path, AddonsRootPath) {
				log.Debugf("addons flow")
				chain.ProcessFilter(req, resp)
				return
			}
			// check external flow token.
			if req.Attribute(SrcPluginIDAttribute) == nil {
				log.Debugf("external flow")
				proxyExternalFlowFilter(externalFilter)(req, resp, chain)
				return
			}
			// check internal flow tkeel version.
			if proxyTkeelVersionFilter(req, resp) {
				chain.ProcessFilter(req, resp)
			}
		}
	}
}

func proxyPluginTokenFilter(req *restful.Request, resp *restful.Response) bool {
	token := req.HeaderParameter(XKeelHeader)
	if token != "" {
		log.Debugf("tkeel plugin token: %s", token)
		s, err := checkToken(token)
		if err != nil {
			log.Errorf("error check token(%s): %s", token, err)
			if err = resp.WriteErrorString(http.StatusForbidden,
				"x-plugin-jwt token invaild"); err != nil {
				log.Errorf("error response write error string(%d,%s): %s",
					http.StatusForbidden, "x-plugin-jwt token invaild", err)
			}
			return false
		}
		req.SetAttribute(SrcPluginIDAttribute, s)
	}
	return true
}

func proxyTkeelVersionFilter(req *restful.Request, resp *restful.Response) bool {
	srcPIDIn := req.Attribute(SrcPluginIDAttribute)
	srcPluginID, ok := srcPIDIn.(string)
	if !ok {
		log.Errorf("error invaild type src plugin id: %v", srcPIDIn)
		resp.WriteErrorString(http.StatusInternalServerError,
			"src plugin type error")
		return false
	}
	if strings.HasPrefix(req.Request.URL.Path, AddonsRootPath) {
		log.Debugf("addons flow")
		return true
	}
	// internal flow.
	sub := getSubPath(req.Request.URL.Path, ApisRootPath)
	dstPluginID := getPluginIDFromApisPath(sub)
	method := req.PathParameter(MethodPath)
	srcPrI, ok := pluginRouteMap.Load(srcPluginID)
	if !ok {
		log.Errorf("error not found src plugin(%s) route", srcPluginID)
		resp.WriteErrorString(http.StatusInternalServerError,
			"not found src plugin("+srcPluginID+")")
		return false
	}
	srcPr, ok := srcPrI.(*model.PluginRoute)
	if !ok {
		log.Errorf("error src pri not plugin route")
		resp.WriteErrorString(http.StatusInternalServerError,
			"internal error")
		return false
	}

	dstPrI, ok := pluginRouteMap.Load(dstPluginID)
	if !ok {
		log.Errorf("error not found dst plugin(%s) route", dstPluginID)
		resp.WriteErrorString(http.StatusInternalServerError,
			"not found dst plugin("+dstPluginID+")")
		return false
	}
	dstPr, ok := dstPrI.(*model.PluginRoute)
	if !ok {
		log.Errorf("error dst pri not plugin route")
		resp.WriteErrorString(http.StatusInternalServerError,
			"internal error")
		return false
	}

	ok, err := util.CheckRegisterPluginTkeelVersion(dstPr.TkeelVersion,
		srcPr.TkeelVersion)
	if err != nil {
		log.Errorf("error check tkeel version(%s/%s): %s",
			dstPr.TkeelVersion, srcPr.TkeelVersion, err)
		resp.WriteErrorString(http.StatusInternalServerError,
			"internal error")
		return false
	}
	if !ok {
		err = fmt.Errorf("error dst plugin tkeel version(%s) not less then src plugin tkeel version(%s)",
			dstPr.TkeelVersion, srcPr.TkeelVersion)
		log.Error(err)
		resp.WriteError(http.StatusForbidden, err)
		return false
	}
	log.Debugf("src(%s) request dst(%s) method(%s)",
		srcPluginID, dstPluginID, method)
	return true
}

func proxyExternalFlowFilter(externalFilter func(*restful.Request, *restful.Response, *restful.FilterChain)) func(*restful.Request, *restful.Response, *restful.FilterChain) {
	return func(req *restful.Request, resp *restful.Response,
		chain *restful.FilterChain) {
		token := req.HeaderParameter(XKeelHeader)
		if token == "" {
			// external flow.
			externalFilter(req, resp, chain)
			return
		}
		chain.ProcessFilter(req, resp)
	}
}
