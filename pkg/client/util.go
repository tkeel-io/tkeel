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

package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/util"
)

const contentTypeJSON = "application/json"

var ErrMethodNotAllow = errors.New("method not allow")

func InvokeJSON(ctx context.Context, c dapr.Client, request *dapr.AppRequest, reqJSON, respJSON interface{}) ([]byte, error) {
	var (
		resp *http.Response
		err  error
	)
	canClose := true
	request.Header.Del("Accept-Encoding")
	if !util.IsNil(reqJSON) && reqJSON != nil {
		request.Header.Set("Content-Type", contentTypeJSON)
		reqBody, err1 := json.Marshal(reqJSON)
		if err1 != nil {
			return nil, errors.Wrapf(err1, "marshal dapr invoke(%s) request", request)
		}
		request.Body = reqBody
		resp, err = c.Call(ctx, request)
		defer func() {
			if err = resp.Body.Close(); err != nil {
				return
			}
		}()
	} else {
		resp, err = c.Call(ctx, request)
		defer func() {
			if canClose {
				if err = resp.Body.Close(); err != nil {
					return
				}
			}
		}()
	}
	if err != nil {
		canClose = false
		return nil, errors.Wrapf(err, "invoke requst(%s)", request)
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusMethodNotAllowed {
			return nil, ErrMethodNotAllow
		}
		return nil, errors.Errorf("error invoke request(%s): %s", request, resp.Status)
	}
	if resp.ContentLength == 0 {
		return nil, nil
	}
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read resp body")
	}
	if !util.IsNil(respJSON) && respJSON != nil {
		err = json.Unmarshal(out, respJSON)
		if err != nil {
			return nil, errors.Wrapf(err, "unmarshal out(%s)", string(out))
		}
	}
	return out, nil
}
