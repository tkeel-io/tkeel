package openapi

import (
	"fmt"
	"net/http"

	"github.com/tkeel-io/tkeel/pkg/logger"
)

var log = logger.NewLogger("keel.openapi")

type Openapi struct {
	port         int
	pluginID     string
	version      string
	requiredFunc struct {
		Identify   func() (*IdentifyResp, error)
		Status     func() (*StatusResp, error)
		TenantBind func(*TenantBindReq) (*TenantBindResp, error)
	}
	optionalFunc struct {
		AddonsIdentify func(*AddonsIdentifyReq) (*AddonsIdentifyResp, error)
	}
	mux            *http.ServeMux
	registerAPIMap map[string]*API

	closeSrv func() error
}

type RequiredFunc struct {
	Identify   func() (*IdentifyResp, error)
	Status     func() (*StatusResp, error)
	TenantBind func(*TenantBindReq) (*TenantBindResp, error)
}

type OptionalFunc struct {
	AddonsIdentify func(*AddonsIdentifyReq) (*AddonsIdentifyResp, error)
}

func NewOpenapi(port int, id, ver string) *Openapi {
	return &Openapi{
		port:           port,
		pluginID:       id,
		version:        ver,
		mux:            http.NewServeMux(),
		registerAPIMap: make(map[string]*API),
	}
}

func (a *Openapi) Listen() error {
	registerRequired(a.mux, a)
	registerOptional(a.mux, a)
	for _, h := range a.registerAPIMap {
		registerHandler(a.mux, h.Endpoint, h.H)
	}
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: a.mux,
	}
	return server.ListenAndServe()
}

func (a *Openapi) Close() error {
	if a.closeSrv != nil {
		return a.closeSrv()
	}
	return nil
}

func (a *Openapi) SetRequiredFunc(r RequiredFunc) {
	a.requiredFunc = r
}

func (a *Openapi) SetOptionalFunc(o OptionalFunc) {
	a.optionalFunc = o
}

func (a *Openapi) AddOpenAPI(o *API) error {
	err := o.Check()
	if err != nil {
		return err
	}
	a.registerAPIMap[o.Endpoint] = o
	return nil
}

// Required API

func (a *Openapi) GetIdentifyResp() *IdentifyResp {
	return &IdentifyResp{
		CommonResult: SuccessResult(),
		PluginID:     a.pluginID,
		Version:      a.version,
	}
}

func (a *Openapi) Identify() (*IdentifyResp, error) {
	if a.requiredFunc.Identify == nil {
		resp := a.GetIdentifyResp()
		return resp, nil
	}
	return a.requiredFunc.Identify()
}

func (a *Openapi) Status() (*StatusResp, error) {
	if a.requiredFunc.Status == nil {
		return &StatusResp{
			CommonResult: SuccessResult(),
			Status:       Active,
		}, nil
	}
	return a.requiredFunc.Status()
}

func (a *Openapi) TenantBind(req *TenantBindReq) (*TenantBindResp, error) {
	if a.requiredFunc.Status == nil {
		return &TenantBindResp{
			CommonResult: SuccessResult(),
		}, nil
	}
	return a.requiredFunc.TenantBind(req)
}

// Optional API

func (a *Openapi) AddonsIdentify(req *AddonsIdentifyReq) (*AddonsIdentifyResp, error) {
	if a.optionalFunc.AddonsIdentify == nil {
		return &AddonsIdentifyResp{
			CommonResult: BadRequestResult("no extension point"),
		}, nil
	}
	return a.optionalFunc.AddonsIdentify(req)
}
