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
	"fmt"

	"github.com/tkeel-io/tkeel/pkg/util"

	dapr "github.com/dapr/go-sdk/client"
)

const contentTypeJSON = "application/json"

func DaprInvokeJSON(ctx context.Context, c dapr.Client, appID, methodName, verb string, req, resp interface{}) error {
	var (
		out []byte
		err error
	)
	if !util.IsNil(req) {
		reqBody, err1 := json.Marshal(req)
		if err1 != nil {
			return fmt.Errorf("error marshal dapr invoke(%s/%s/%s) request: %w", appID, methodName, verb, err1)
		}
		out, err = c.InvokeMethodWithContent(ctx, appID, methodName, verb, &dapr.DataContent{
			Data:        reqBody,
			ContentType: contentTypeJSON,
		})
	}

	if err != nil {
		return fmt.Errorf("error invoke app(%s) method(%s/%s): %w", appID, methodName, verb, err)
	}

	if !util.IsNil(resp) {
		err = json.Unmarshal(out, resp)
		if err != nil {
			return fmt.Errorf("error unmarshal app(%s) method(%s/%s): %w", appID, methodName, verb, err)
		}
	}
	return nil
}
