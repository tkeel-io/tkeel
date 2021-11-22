/*
Copyright 2021 The tKeel Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
	http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dapr

import (
	"context"
	"net/http"
	"net/url"
)

type AppReqeust struct {
	ID         string      `json:"id"`
	Method     string      `json:"method"`
	Verb       string      `json:"verb"`
	Header     http.Header `json:"header"`
	QueryValue url.Values  `json:"query_value"`
	Body       []byte      `json:"body"`
}

type Client interface {
	Call(context.Context, *AppReqeust) (*http.Response, error)
}

type HTTPClient struct {
	httpAddr string
}

func NewHTTPClient(httpPort string) *HTTPClient {
	return &HTTPClient{
		httpAddr: "127.0.0.1:" + httpPort,
	}
}
