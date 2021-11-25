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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/tkeel-io/kit/log"

	dapr "github.com/dapr/go-sdk/client"
)

type AppRequest struct {
	ID         string      `json:"id"`
	Method     string      `json:"method"`
	Verb       string      `json:"verb"`
	Header     http.Header `json:"header"`
	QueryValue url.Values  `json:"query_value"`
	Body       []byte      `json:"body"`
}

func (a *AppRequest) String() string {
	b, err := json.Marshal(a)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

type Client interface {
	Call(context.Context, *AppRequest) (*http.Response, error)
}

type HTTPClient struct {
	httpAddr string
}

func NewHTTPClient(httpPort string) *HTTPClient {
	return &HTTPClient{
		httpAddr: "127.0.0.1:" + httpPort,
	}
}

func NewGPRCClient(retry int, interval, gprcPort string) (dapr.Client, error) {
	var daprGRPCClient dapr.Client
	var err error
	inval, err := time.ParseDuration(interval)
	if err != nil {
		return nil, fmt.Errorf("error parse interval(%s): %w", interval, err)
	}
	if retry < 1 {
		retry = 1
	}
	for i := 0; i < retry; i++ {
		daprGRPCClient, err = dapr.NewClientWithPort(gprcPort)
		if err == nil {
			break
		}

		time.Sleep(inval)
		log.Debugf("error new client: %s retry: %d", err, i)
	}
	if err != nil {
		return nil, fmt.Errorf("error new client with port(%s): %w", gprcPort, err)
	}
	return daprGRPCClient, nil
}
