package openapi

import (
	"encoding/json"
	"fmt"
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
					return nil, fmt.Errorf("error identify: %w", err)
				}

				respByte, err := json.Marshal(resp)
				if err != nil {
					return nil, fmt.Errorf("error json marshal: %w", err)
				}
				return respByte, nil
			}))
	registerHandler(mux, "/v1/status",
		convertFunc2Handler(http.MethodGet,
			func(b []byte) ([]byte, error) {
				resp, err := apiRequred.Status()
				if err != nil {
					return nil, fmt.Errorf("error status: %w", err)
				}

				respByte, err := json.Marshal(resp)
				if err != nil {
					return nil, fmt.Errorf("error json marshal: %w", err)
				}
				return respByte, nil
			}))
	registerHandler(mux, "/v1/tenant/bind",
		convertFunc2Handler(http.MethodPost,
			func(b []byte) ([]byte, error) {
				req := &TenantBindReq{}
				err := json.Unmarshal(b, req)
				if err != nil {
					return nil, fmt.Errorf("error json unmarshal: %w", err)
				}
				resp, err := apiRequred.TenantBind(req)
				if err != nil {
					return nil, fmt.Errorf("error tenant bind: %w", err)
				}

				respByte, err := json.Marshal(resp)
				if err != nil {
					return nil, fmt.Errorf("error json marshal: %w", err)
				}
				return respByte, nil
			}))
}
