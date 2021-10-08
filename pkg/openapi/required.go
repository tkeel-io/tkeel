package openapi

import (
	"encoding/json"
	"net/http"
)

type Required interface {
	Identify() (*IdentifyResp, error)
	Status() (*StatusResp, error)
	TenantBind(*TenantBindReq) (*TenantBindResp, error)
}

func registerRequired(mux *http.ServeMux, apiRequred Required) {
	registerHandler(mux, "/v1/identify",
		convertFunc2Handler(http.MethodGet,
			func(b []byte) ([]byte, error) {
				resp, err := apiRequred.Identify()
				if err != nil {
					return nil, err
				}

				respByte, err := json.Marshal(resp)
				if err != nil {
					return nil, err
				}
				return respByte, nil
			}))
	registerHandler(mux, "/v1/status",
		convertFunc2Handler(http.MethodGet,
			func(b []byte) ([]byte, error) {
				resp, err := apiRequred.Status()
				if err != nil {
					return nil, err
				}

				respByte, err := json.Marshal(resp)
				if err != nil {
					return nil, err
				}
				return respByte, nil
			}))
	registerHandler(mux, "/v1/tenant/bind",
		convertFunc2Handler(http.MethodPost,
			func(b []byte) ([]byte, error) {
				req := &TenantBindReq{}
				err := json.Unmarshal(b, req)
				if err != nil {
					return nil, err
				}
				resp, err := apiRequred.TenantBind(req)
				if err != nil {
					return nil, err
				}

				respByte, err := json.Marshal(resp)
				if err != nil {
					return nil, err
				}
				return respByte, nil
			}))
}
