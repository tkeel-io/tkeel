package keel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/utils"
)

func auth(e *openapi.APIEvent) (string, error) {
	if ok := e.HTTPReq.Header.Get("x-keel"); ok != "" {
		log.Debugf("self request")
		return "keel", nil
	}
	// get plugin id.
	pID, err := keel.GetPluginIDFromRequest(e.HTTPReq)
	if err != nil {
		return pID, fmt.Errorf("error get plugin id from request: %w", err)
	}
	if pID != "" {
		log.Debug("internal flow")
		return pID, nil
	}
	if err := externalPreRouteCheck(e.HTTPReq.Context(), e.HTTPReq); err != nil {
		return pID, fmt.Errorf("error external pre route check: %w", err)
	}
	return pID, nil
}

func getAddonsUpstream(ctx context.Context,
	pID, path string) (upstreamPath string, err error) {
	addonsPoint := strings.TrimPrefix(path, keel.AddonsURLPrefix)
	route, _, err := keel.GetPluginRoute(ctx, pID)
	if err != nil {
		return "", fmt.Errorf("error get plugin route: %w", err)
	}
	if route.RegisterAddons == nil {
		return "", nil
	}
	upstreamPath, ok := route.RegisterAddons[addonsPoint]
	if !ok {
		return "", nil
	}
	return upstreamPath, nil
}

func getUpstreamPlugin(ctx context.Context, pID, path string) (string, string, error) {
	var upstreamPath string
	if strings.HasPrefix(path, "/"+keel.AddonsPath) {
		if pID == "" {
			return "", "", errors.New("request addons not internal flow")
		}
		up, err := getAddonsUpstream(ctx, pID, path)
		if err != nil {
			return "", "", fmt.Errorf("error get addons upstream: %w", err)
		}
		upstreamPath = up
	} else {
		upstreamPath = strings.TrimPrefix(path, "/")
	}

	if upstreamPath == "" {
		log.Debugf("not found registered addons: %s %s", pID, path)
		return "", "", errors.New("not found")
	}
	upPluginID, endpoint := keel.DecodeRoute(upstreamPath)
	if upPluginID == "" || endpoint == "" {
		return "", "", fmt.Errorf("error request %s", upstreamPath)
	}
	return upPluginID, endpoint, nil
}

func checkPluginStatus(ctx context.Context, pID string) error {
	route, _, err := keel.GetPluginRoute(ctx, pID)
	if err != nil {
		return fmt.Errorf("error get plugin route: %w", err)
	}
	if route == nil {
		return errors.New("not registered")
	}
	if route.Status != openapi.Active && route.Status != openapi.Starting {
		return fmt.Errorf("%s not ACTIVE or STARTING", route.Status)
	}
	return nil
}

func externalPreRouteCheck(ctx context.Context, req *http.Request) error {
	reqHeader := req.Header.Clone()
	reqHeader.Add("x-keel", "True")
	resp, err := keel.CallAddons(ctx, "externalPreRouteCheck", req.Method, &keel.CallReq{
		Header:   reqHeader,
		URLValue: req.URL.Query(),
		Body: []byte(`Check the request header information,
		return a status code 200,body: 
		{
			"msg":"ok", // costom msg
			"ret":0, // must be zero
		}
		If invalid, return a status code other than 200 or return body:
		{
			"msg":"faild", // costom msg
			"ret":-1, // negative
		}`),
	})
	if err != nil {
		return fmt.Errorf("error call addons: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Errorf("error response body close: %s", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			log.Debugf("not found registered addons point(externalPreRouteCheck)")
			return nil
		}
		return errors.New(resp.Status)
	}
	result := &openapi.CommonResult{}
	if err := utils.ReadBody2Json(resp.Body, result); err != nil {
		log.Errorf("error read externalPreRouteCheck func: %w", err)
		return fmt.Errorf("error read externalPreRouteCheck func: %w", err)
	}
	if result.Ret != 0 {
		log.Errorf("error externalPreRouteCheck: %s", result.Msg)
		return errors.New(result.Msg)
	}
	return nil
}

// check Route path vaild return continue and error.
func checkRoutePath(path string) (bool, error) {
	// check path.
	if path == "/" {
		return false, errors.New("error invaild path: /")
	}
	if strings.HasPrefix(path, "/dapr") {
		log.Debugf("dapr sidecar request: %s", path)
		return false, nil
	}
	return true, nil
}

func proxy(req *http.Request, respWrite http.ResponseWriter, pluginID, endpoint string) (*http.Response, error) {
	// request.
	bodyByte, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("error get request body: %w", err)
	}
	defer req.Body.Close()

	resp, err := keel.CallPlugin(req.Context(), pluginID, endpoint,
		req.Method, &keel.CallReq{
			Header:   req.Header.Clone(),
			Body:     bodyByte,
			URLValue: req.URL.Query(),
		})
	if err != nil {
		if resp != nil {
			http.Error(respWrite, resp.Status, resp.StatusCode)
		} else {
			http.Error(respWrite, "bad request", http.StatusBadRequest)
		}
		return nil, fmt.Errorf("error request(%s/%s/%s) : %w", pluginID, endpoint,
			req.Method, err)
	}
	log.Debugf("req(%s/%s) body(%s) -->resp(%s)",
		pluginID, endpoint, bodyByte, resp.Status)
	return resp, nil
}

func copyHeader(dst, src http.Header) {
	for k, values := range src {
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}
