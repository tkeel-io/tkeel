package keel

import (
	"io"
	"net/http"
	"strings"

	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/openapi"
)

func (g *Keel) auth(e *openapi.APIEvent, pID string) error {
	if pID != "" {
		log.Debug("internal flow")
		return nil
	}
	return externalPreRouteCheck(e.HttpReq.Context(), e.HttpReq)
}

func (g *Keel) Route(e *openapi.APIEvent) {
	// check path
	path := e.HttpReq.RequestURI
	if path == "/" {
		log.Error("error invaild path: /")
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}
	if strings.HasPrefix(path, "/dapr") {
		log.Debugf("dapr sidecar request: %s", e.HttpReq.RequestURI)
		http.Error(e, "not found", http.StatusNotFound)
		return
	}

	// get plugin id
	pID, err := keel.GetPluginIDFromRequest(e.HttpReq)
	if err != nil {
		log.Errorf("error get plugin id from request: %s", err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}

	log.Debugf("route pID(%s) requst %s", pID, e.HttpReq.RequestURI)

	if ok := e.HttpReq.Header.Get("x-keel"); ok != "" {
		log.Debugf("self request")
		pID = "keel"
	} else {
		// auth check
		err = g.auth(e, pID)
		if err != nil {
			log.Errorf("error auth: %s", err)
			http.Error(e, "auth faild", http.StatusUnauthorized)
			return
		}
	}

	// find upstream plugin
	upstreamPath, err := getUpstreamPath(e.HttpReq.Context(), pID, path)
	if err != nil {
		log.Errorf("error get upstream path: %s", err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}
	if upstreamPath == "" {
		log.Debugf("not found registered addons: %s %s", pID, path)
		http.Error(e, "not found", http.StatusNotFound)
		return
	}
	upPluginID, endpoint := keel.DecodeRoute(upstreamPath)
	if upPluginID == "" || endpoint == "" {
		log.Errorf("error request %s", upstreamPath)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}

	// check upstream plugin
	err = checkPluginStatus(e.HttpReq.Context(), upPluginID)
	if err != nil {
		log.Errorf("error check plugin(%s) status: %s", upPluginID, err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}

	// request
	bodyByte, err := io.ReadAll(e.HttpReq.Body)
	if err != nil {
		log.Errorf("error get request body: %s", err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}
	defer e.HttpReq.Body.Close()

	resp, err := keel.CallPlugin(e.HttpReq.Context(), upPluginID, endpoint,
		e.HttpReq.Method, &keel.CallReq{
			Header:   e.HttpReq.Header.Clone(),
			Body:     bodyByte,
			UrlValue: e.HttpReq.URL.Query(),
		})
	if err != nil {
		log.Errorf("error request(%s/%s/%s) : %s", upPluginID, endpoint,
			e.HttpReq.Method, err)
		if resp != nil {
			http.Error(e, resp.Status, resp.StatusCode)
		} else {
			http.Error(e, "bad request", http.StatusBadRequest)
		}
		return
	}
	defer resp.Body.Close()

	for k, values := range resp.Header {
		if k == "x-plugin-jwt" {
			// 插件间调用 ==>插件调用平台+平台调用插件
			continue
		}
		for _, v := range values {
			e.Header().Add(k, v)
		}
	}
	respBodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("error get response body: %s", err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}
	_, err = e.Write(respBodyByte)
	if err != nil {
		log.Errorf("error response write: %s", err)
	}

	log.Debugf("plugin(%s/%s) req(%s) --> plugin(%s/%s) resp(%s)",
		pID, path, string(bodyByte), upPluginID, endpoint, string(respBodyByte))
}
