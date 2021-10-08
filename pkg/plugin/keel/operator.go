package keel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/utils"
)

func getAddonsUpstream(ctx context.Context,
	pID, path string) (upstreamPath string, err error) {
	addonsPoint := strings.TrimPrefix(path, keel.ADDONS_URL_PRIFIX)
	route, _, err := keel.GetPluginRoute(ctx, pID)
	if err != nil {
		return "", err
	}
	upstreamPath, ok := route.Addons[addonsPoint]
	if !ok {
		return "", nil
	}
	return upstreamPath, nil
}

func getUpstreamPath(ctx context.Context, pID, path string) (string, error) {
	if strings.HasPrefix(path, "/"+keel.ADDONS_PATH) {
		if pID == "" {
			return "", errors.New("request addons not internal flow")
		}
		up, err := getAddonsUpstream(ctx, pID, path)
		if err != nil {
			return "", err
		}
		return up, nil
	}
	return strings.TrimPrefix(path, "/"), nil
}

func checkPluginStatus(ctx context.Context, pID string) error {
	route, _, err := keel.GetPluginRoute(ctx, pID)
	if err != nil {
		return err
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
		UrlValue: req.URL.Query(),
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
		return err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			log.Debugf("not found registered addons point(externalPreRouteCheck)")
			return nil
		}
		return errors.New(resp.Status)
	}
	result := &openapi.CommonResult{}
	if err := utils.ReadBody2Json(resp.Body, result); err != nil {
		log.Errorf("error read externalPreRouteCheck func: %s", err)
		return err
	}
	if result.Ret != 0 {
		log.Errorf("error externalPreRouteCheck: %s", result.Msg)
		return errors.New(result.Msg)
	}
	return nil
}
