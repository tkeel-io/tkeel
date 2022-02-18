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
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const daprInvokeURLTemplate = "http://%s/v1.0/invoke/%s/method/%s"

// Call http invake dapr sidecar.
func (c *HTTPClient) Call(ctx context.Context,
	req *AppRequest) (*http.Response, error) {
	url := c.getInvokeURL(req)
	if len(req.QueryValue) != 0 {
		url += "?" + req.QueryValue.Encode()
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Verb, url, bytes.NewReader(req.Body))
	if err != nil {
		return nil, errors.Wrap(err, "http new request")
	}
	if len(req.Header) != 0 {
		httpReq.Header = req.Header.Clone()
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "error http default client do")
	}
	return resp, nil
}

func (c *HTTPClient) getInvokeURL(req *AppRequest) string {
	return fmt.Sprintf(daprInvokeURLTemplate, c.httpAddr, req.ID, req.Method)
}
