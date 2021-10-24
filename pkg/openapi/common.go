package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	APIIdentifyMethod       = "v1/identify"
	APIStatusMethod         = "v1/status"
	APIAddonsIdentifyMethod = "v1/addons/identify"
	APITenantBindMethod     = "v1/tenant/bind"
)

// AddonsPoint Plugin extension point.
type AddonsPoint struct {
	AddonsPoint string `json:"addons_point"`
	Desc        string `json:"desc,omitempty"`
}

type AddonsEndpoint struct {
	AddonsPoint string `json:"addons_point"`
	Endpoint    string `json:"endpoint"`
}

// MainPlugin Main plugin for plugin extension.
type MainPlugin struct {
	ID        string            `json:"id"`
	Version   string            `json:"version,omitempty"`
	Endpoints []*AddonsEndpoint `json:"endpoints"`
}

// IdentifyResp response of /v1/identify.
type IdentifyResp struct {
	CommonResult `json:",inline"`
	PluginID     string         `json:"plugin_id"`
	Version      string         `json:"version"`
	TkeelVersion string         `json:"tkeel_version"`
	AddonsPoints []*AddonsPoint `json:"addons_points,omitempty"`
	MainPlugins  []*MainPlugin  `json:"main_plugins,omitempty"`
}

// AddonsIdentifyReq request of /v1/addons/idnetify POST.
type AddonsIdentifyReq struct {
	Plugin struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	} `json:"plugin"`
	Endpoint []*AddonsEndpoint `json:"endpoint"`
}

// AddonsIdentifyResp response of /v1/addons/identify.
type AddonsIdentifyResp struct {
	CommonResult `json:",inline"`
}

// TenantBindReq request of /v1/tenant/bind POST.
type TenantBindReq struct {
	TenantID string `json:"tenant_id"`
	Extra    []byte `json:"extra,omitempty"`
}

// TenantBindResp response of /v1/tenant/bind.
type TenantBindResp struct {
	CommonResult `json:",inline"`
}

// PluginStatus plugin status.
type PluginStatus string

const (
	Starting PluginStatus = "STARTING"
	Active   PluginStatus = "ACTIVE"
	Stopping PluginStatus = "STOPPING"
	Stopped  PluginStatus = "STOPPED"
	Failed   PluginStatus = "FAILED"
)

// StatusResp response of /v1/Status.
type StatusResp struct {
	CommonResult `json:",inline"`
	Status       PluginStatus `json:"status"`
}

// CommonResult open api request common response.
type CommonResult struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}

func SuccessResult() CommonResult {
	return CommonResult{
		Ret: 0,
		Msg: "ok",
	}
}

func BadRequestResult(msg string) CommonResult {
	return CommonResult{
		Ret: -400,
		Msg: msg,
	}
}

func InternalErrorResult(msg string) CommonResult {
	return CommonResult{
		Ret: -500,
		Msg: msg,
	}
}

type Endpoint struct {
	KeelMethod string
	Endpoint   string
}

type APIEvent struct {
	http.ResponseWriter
	HTTPReq *http.Request
}

// Handler need return http resp.
type Handler func(*APIEvent)

type Desc struct {
	Verb     []string
	Request  string
	Response string
	Desc     string
}

type API struct {
	Endpoint string
	H        Handler
}

func (a *API) Check() error {
	if a.Endpoint == "" {
		return errors.New("not define endpoint or method")
	}
	if !strings.HasPrefix(a.Endpoint, "/") {
		return fmt.Errorf("endpoint is invaild: %s", a.Endpoint)
	}
	return nil
}

func registerHandler(mux *http.ServeMux, path string, h Handler) {
	mux.HandleFunc(path, func(rw http.ResponseWriter, r *http.Request) {
		for k, values := range r.Header {
			if k == "Authorization" {
				for _, v := range values {
					rw.Header().Add(k, v)
				}
			}
			if k == "x-plugin-jwt" {
				log.Debugf("plugin jwt: %s", values[0])
			}
		}
		event := &APIEvent{
			ResponseWriter: rw,
			HTTPReq:        r,
		}
		h(event)
	})
}

func convertFunc2Handler(matchMethod string, f func([]byte) ([]byte, error)) Handler {
	return func(a *APIEvent) {
		if a.HTTPReq.Method != matchMethod {
			log.Errorf("not support method: %s", a.HTTPReq.Method)
			http.Error(a, http.ErrNotSupported.ErrorString, http.StatusMethodNotAllowed)
			return
		}
		var content []byte

		if matchMethod != http.MethodGet {
			// check for post with no data.
			if a.HTTPReq.ContentLength <= 0 {
				log.Error("content cannot be empty.")
				http.Error(a, "content cannot be empty", http.StatusBadRequest)
				return
			}
			// read content.
			if a.HTTPReq.Close {
				log.Error("request has been closed.")
				http.Error(a, "request has been closed", http.StatusBadRequest)
				return
			}
			readByte, err := ioutil.ReadAll(a.HTTPReq.Body)
			if err != nil {
				log.Errorf("error read body: %s", err)
				http.Error(a, err.Error(), http.StatusBadRequest)
				return
			}
			defer a.HTTPReq.Body.Close()
			content = readByte
		}

		resp, err := f(content)
		if err != nil {
			log.Error(err.Error())
			http.Error(a, err.Error(), http.StatusInternalServerError)
			return
		}

		a.Header().Set("Content-type", "application/json")
		if _, err := a.Write(resp); err != nil {
			http.Error(a, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (a *APIEvent) ResponseJSON(res interface{}) []byte {
	bRes, _ := json.Marshal(res)
	a.Header().Set("Content-Type", "application/json")
	a.Write(bRes)
	return bRes
}
