package openapi

import (
	"encoding/json"
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
					return nil, err
				}

				resp, err := apiOptional.AddonsIdentify(req)
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
