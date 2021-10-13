package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Optional interface {
	AddonsIdentify(*AddonsIdentifyReq) (*AddonsIdentifyResp, error)
}

func registerOptional(mux *http.ServeMux, apiOptional Optional) {
	registerHandler(mux, "/v1/addons/identify",
		convertFunc2Handler(http.MethodPost,
			func(b []byte) ([]byte, error) {
				req := &AddonsIdentifyReq{}
				err := json.Unmarshal(b, req)
				if err != nil {
					return nil, fmt.Errorf("error json unmashal: %w", err)
				}

				resp, err := apiOptional.AddonsIdentify(req)
				if err != nil {
					return nil, fmt.Errorf("error addons identify: %w", err)
				}

				respByte, err := json.Marshal(resp)
				if err != nil {
					return nil, fmt.Errorf("error json marshal: %w", err)
				}
				return respByte, nil
			}))
}
