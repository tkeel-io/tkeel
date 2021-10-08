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
	IDENTIFY_METHOD       = "v1/identify"
	STATUS_METHOD         = "v1/status"
	ADDONSIDENTIFY_METHOD = "v1/addons/identify"
	TENANTBIND_METHOD     = "v1/tenant/bind"
)

// AddonsPoint Plugin extension point
type AddonsPoint struct {
	AddonsPoint string `json:"addons_point"`
	Desc        string `json:"desc,omitempty"`
}

type AddonsEndpoint struct {
	AddonsPoint string `json:"addons_point"`
	Endpoint    string `json:"endpoint"`
}

// MainPlugin Main plugin for plugin extension
type MainPlugin struct {
	ID        string            `json:"id"`
	Version   string            `json:"version,omitempty"`
	Endpoints []*AddonsEndpoint `json:"endpoints"`
}

// IdentifyResp response of /v1/identify
type IdentifyResp struct {
	CommonResult `json:",inline"`
	PluginID     string         `json:"plugin_id"`
	Version      string         `json:"version"`
	AddonsPoints []*AddonsPoint `json:"addons_points,omitempty"`
	MainPlugins  []*MainPlugin  `json:"main_plugins,omitempty"`
}

// AddonsIdentifyReq request of /v1/addons/idnetify POST
type AddonsIdentifyReq struct {
	Plugin struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	} `json:"plugin"`
	Endpoint []*AddonsEndpoint `json:"endpoint"`
}

// AddonsIdentifyResp response of /v1/addons/identify
type AddonsIdentifyResp struct {
	CommonResult `json:",inline"`
}

// TenantBindReq request of /v1/tenant/bind POST
type TenantBindReq struct {
	TenantID string `json:"tenant_id"`
	Extra    []byte `json:"extra"`
}

// TenantBindResp response of /v1/tenant/bind
type TenantBindResp struct {
	CommonResult `json:",inline"`
}

// PluginStatus plugin status
type PluginStatus string

const (
	Starting PluginStatus = "STARTING"
	Active   PluginStatus = "ACTIVE"
	Stopping PluginStatus = "STOPPING"
	Stopped  PluginStatus = "STOPPED"
	Failed   PluginStatus = "FAILED"
)

// StatusResp response of /v1/Status
type StatusResp struct {
	CommonResult `json:",inline"`
	Status       PluginStatus `json:"status"`
}

// CommonResult open api request common response
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
	HttpReq *http.Request
}

// Handler need return http resp
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

func (o *API) Check() error {
	if o.Endpoint == "" {
		return errors.New("not define endpoint or method")
	}
	if !strings.HasPrefix(o.Endpoint, "/") {
		return fmt.Errorf("endpoint is invaild: %s", o.Endpoint)
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
			HttpReq:        r,
		}
		h(event)
	})
}

func convertFunc2Handler(matchMethod string, Func func([]byte) ([]byte, error)) Handler {
	return func(e *APIEvent) {
		if e.HttpReq.Method != matchMethod {
			log.Errorf("not support method: %s", e.HttpReq.Method)
			http.Error(e, http.ErrNotSupported.ErrorString, http.StatusMethodNotAllowed)
			return
		}
		var content []byte
		var err error

		if matchMethod != http.MethodGet {
			// check for post with no data
			if e.HttpReq.ContentLength <= 0 {
				log.Error("content cannot be empty")
				http.Error(e, "content cannot be empty", http.StatusBadRequest)
				return
			}
			// read content
			content, err = ioutil.ReadAll(e.HttpReq.Body)
			if err != nil {
				log.Error(err.Error())
				http.Error(e, err.Error(), http.StatusBadRequest)
				return
			}
		}

		resp, err := Func(content)
		if err != nil {
			log.Error(err.Error())
			http.Error(e, err.Error(), http.StatusInternalServerError)
			return
		}

		e.Header().Set("Content-type", "application/json")
		if _, err := e.Write(resp); err != nil {
			http.Error(e, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (e *APIEvent) ResponseJson(res interface{}) []byte {
	bRes, _ := json.Marshal(res)
	e.Header().Set("Content-Type", "application/json")
	e.Write(bRes)
	return bRes
}
