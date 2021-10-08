package keel

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func EncodeRoute(pID, endpoint string) string {
	return fmt.Sprintf("%s/%s", pID, endpoint)
}

func DecodeRoute(path string) (pluginID, endpoint string) {
	sub := strings.SplitN(path, "/", 2)
	if len(sub) != 2 {
		return "", ""
	}
	sub1 := strings.Split(sub[1], "?")
	return sub[0], sub1[0]
}

func GetDaprInvokeUrl(appID, method string) string {
	return fmt.Sprintf(PLUGIN_INVOKE_URL, daprAddr, appID, method)
}

func GetPluginMethodUrl(pluginID, method string) string {
	return fmt.Sprintf("%s/%s", pluginID, method)
}

func GetAddonsUrl(addonsPoint string) string {
	return fmt.Sprintf("%s/%s", ADDONS_PATH, addonsPoint)
}

func CallPlugin(ctx context.Context, pluginID, method, httpMethod string, req *CallReq) (*http.Response, error) {
	var (
		err     error
		httpReq *http.Request
	)
	invokeUrl := GetDaprInvokeUrl(pluginID, method)
	if req != nil {
		if len(req.UrlValue) != 0 {
			invokeUrl += "?" + req.UrlValue.Encode()
		}
		log.Debugf("call plugins url(%s)", invokeUrl)
		httpReq, err = http.NewRequest(httpMethod, invokeUrl, bytes.NewReader(req.Body))
		if err != nil {
			return nil, err
		}
		if req.Header != nil {
			httpReq.Header = req.Header.Clone()
		}
	} else {
		log.Debugf("call plugins url(%s)", invokeUrl)
		httpReq, err = http.NewRequest(httpMethod, invokeUrl, nil)
		if err != nil {
			return nil, err
		}
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	return resp, nil
}

func CallKeel(ctx context.Context, pluginID, method, httpMethod string, req *CallReq) (*http.Response, error) {
	var (
		err     error
		httpReq *http.Request
	)
	invokeUrl := GetDaprInvokeUrl("keel", GetPluginMethodUrl(pluginID, method))
	if req != nil {
		if len(req.UrlValue) != 0 {
			invokeUrl += "?" + req.UrlValue.Encode()
		}
		log.Debugf("call keel url(%s)", invokeUrl)
		httpReq, err = http.NewRequest(httpMethod, invokeUrl, bytes.NewReader(req.Body))
		if req.Header != nil {
			httpReq.Header = req.Header.Clone()
		}
	} else {
		log.Debugf("call keel url(%s)", invokeUrl)
		httpReq, err = http.NewRequest(httpMethod, invokeUrl, nil)
	}
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func CallAddons(ctx context.Context, addonsPoint, httpMethod string, req *CallReq) (*http.Response, error) {
	var (
		err     error
		httpReq *http.Request
	)
	invokeUrl := GetDaprInvokeUrl("keel", GetAddonsUrl(addonsPoint))
	if req != nil {
		if len(req.UrlValue) != 0 {
			invokeUrl += "?" + req.UrlValue.Encode()
		}
		log.Debugf("call addons url(%s)", invokeUrl)
		httpReq, err = http.NewRequest(httpMethod, invokeUrl, bytes.NewReader(req.Body))
		if req.Header != nil {
			httpReq.Header = req.Header.Clone()
		}
	} else {
		log.Debugf("call addons url(%s)", invokeUrl)
		httpReq, err = http.NewRequest(httpMethod, invokeUrl, nil)
	}
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
