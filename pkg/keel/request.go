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

func GetDaprInvokeURL(appID, method string) string {
	return fmt.Sprintf(PluginInvokeURLFormat, daprAddr, appID, method)
}

func GetPluginMethodURL(pluginID, method string) string {
	return fmt.Sprintf("%s/%s", pluginID, method)
}

func GetAddonsURL(addonsPoint string) string {
	return fmt.Sprintf("%s/%s", AddonsPath, addonsPoint)
}

func CallPlugin(ctx context.Context, pluginID, method, httpMethod string, req *CallReq) (*http.Response, error) {
	var (
		err     error
		httpReq *http.Request
	)
	invokeURL := GetDaprInvokeURL(pluginID, method)
	if req != nil {
		if len(req.URLValue) != 0 {
			invokeURL += "?" + req.URLValue.Encode()
		}
		log.Debugf("call plugins url(%s)", invokeURL)
		httpReq, err = http.NewRequest(httpMethod, invokeURL, bytes.NewReader(req.Body))
		if err != nil {
			return nil, fmt.Errorf("error http request: %w", err)
		}
		if req.Header != nil {
			httpReq.Header = req.Header.Clone()
		}
	} else {
		log.Debugf("call plugins url(%s)", invokeURL)
		httpReq, err = http.NewRequest(httpMethod, invokeURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error http request: %w", err)
		}
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error http client do: %w", err)
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
	invokeURL := GetDaprInvokeURL("keel", GetPluginMethodURL(pluginID, method))
	if req != nil {
		if len(req.URLValue) != 0 {
			invokeURL += "?" + req.URLValue.Encode()
		}
		log.Debugf("call keel url(%s)", invokeURL)
		httpReq, err = http.NewRequest(httpMethod, invokeURL, bytes.NewReader(req.Body))
		if req.Header != nil {
			httpReq.Header = req.Header.Clone()
		}
	} else {
		log.Debugf("call keel url(%s)", invokeURL)
		httpReq, err = http.NewRequest(httpMethod, invokeURL, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("error http request: %w", err)
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error http client do: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error http request: (%d/%s)", resp.StatusCode, resp.Status)
	}
	return resp, nil
}

func CallAddons(ctx context.Context, addonsPoint, httpMethod string, req *CallReq) (*http.Response, error) {
	var (
		err     error
		httpReq *http.Request
	)
	invokeURL := GetDaprInvokeURL("keel", GetAddonsURL(addonsPoint))
	if req != nil {
		if len(req.URLValue) != 0 {
			invokeURL += "?" + req.URLValue.Encode()
		}
		log.Debugf("call addons url(%s)", invokeURL)
		httpReq, err = http.NewRequest(httpMethod, invokeURL, bytes.NewReader(req.Body))
		if req.Header != nil {
			httpReq.Header = req.Header.Clone()
		}
	} else {
		log.Debugf("call addons url(%s)", invokeURL)
		httpReq, err = http.NewRequest(httpMethod, invokeURL, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("error http request: %w", err)
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error http client do: %w", err)
	}
	return resp, nil
}
