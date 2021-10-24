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

var (
	ErrNotFound        = errors.New("not found")
	ErrNotInternalFlow = errors.New("request addons not internal flow")
	ErrNotRegister     = errors.New("not registered")
	ErrInvaildPath     = errors.New("error invaild path: /")
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
	ps := strings.Split(path, "?")
	addonsPoint := strings.TrimPrefix(ps[0], keel.AddonsURLPrefix)
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

func getUpstreamPlugin(ctx context.Context, pID, path string) (string, string, string, error) {
	var upstreamPath string
	var requestTkeelVersion string
	if strings.HasPrefix(path, "/"+keel.AddonsPath) {
		if pID == "" {
			return "", "", "", errors.New("request addons not internal flow")
		}
		up, err := getAddonsUpstream(ctx, pID, path)
		if err != nil {
			return "", "", "", fmt.Errorf("error get addons upstream: %w", err)
		}
		upstreamPath = up
	} else {
		upstreamPath = strings.TrimPrefix(path, "/")
		if pID == "" {
			u := strings.SplitN(upstreamPath, "/", 2)
			if len(u) != 2 {
				return "", "", "", fmt.Errorf("error external request path: %s", upstreamPath)
			}
			upstreamPath = u[1]
			requestTkeelVersion = u[0]
		}
	}

	if upstreamPath == "" {
		log.Debugf("not found registered addons: %s %s", pID, path)
		return "", "", "", ErrNotFound
	}

	log.Debugf("upstreamPath: %s", upstreamPath)
	upPluginID, endpoint := keel.DecodeRoute(upstreamPath)
	if upPluginID == "" || endpoint == "" {
		return "", "", "", fmt.Errorf("error request %s", upstreamPath)
	}
	return requestTkeelVersion, upPluginID, endpoint, nil
}

func checkDependVersion(src, curr, up string) error {
	log.Debugf("src(%v) cur(%v) up(%v)", src, curr, up)
	ok, err := keel.CheckRegisterPluginTkeelVersion(src, curr)
	if err != nil {
		return fmt.Errorf("error check src tkeel version: %w", err)
	}
	if !ok {
		return fmt.Errorf("error src(%s) > curr(%s)",
			src, curr)
	}
	ok, err = keel.CheckRegisterPluginTkeelVersion(src, up)
	if err != nil {
		return fmt.Errorf("error check up tkeel version: %w", err)
	}
	if !ok {
		return fmt.Errorf("error src(%s) > up(%s)",
			src, up)
	}
	return nil
}

func checkUpstreamPlugin(ctx context.Context, srcPID, upPID, exTVer, curTVer string) error {
	srcDependVersion := exTVer
	if srcPID != "" {
		route, _, err := keel.GetPluginRoute(ctx, srcPID)
		if err != nil {
			return fmt.Errorf("error get src plugin route: %w", err)
		}
		srcDependVersion = route.TkeelVersion
	}
	upRoute, _, err := keel.GetPluginRoute(ctx, upPID)
	if err != nil {
		return fmt.Errorf("error get up plugin route: %w", err)
	}
	if upRoute == nil {
		return ErrNotRegister
	}

	err = checkDependVersion(srcDependVersion, curTVer, upRoute.TkeelVersion)
	if err != nil {
		return fmt.Errorf("error check depend version: %w", err)
	}

	if upRoute.Status != openapi.Active && upRoute.Status != openapi.Starting {
		return fmt.Errorf("%s not ACTIVE or STARTING", upRoute.Status)
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
		return fmt.Errorf("error read externalPreRouteCheck func: %w", err)
	}
	if result.Ret != 0 {
		return errors.New(result.Msg)
	}
	return nil
}

// check Route path vaild return continue and error.
func checkRoutePath(path string) (bool, error) {
	// check path.
	if path == "/" {
		return false, ErrNotRegister
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
