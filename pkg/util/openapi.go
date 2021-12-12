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

package util

import v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"

func GetV1ResultOK() *v1.Result {
	return &v1.Result{
		Ret: v1.Retcode_OK,
		Msg: "ok",
	}
}

func GetV1ResultBadRequest(msg string) *v1.Result {
	return &v1.Result{
		Ret: v1.Retcode_BAD_REQEUST,
		Msg: msg,
	}
}

func GetV1ResultInternalError(msg string) *v1.Result {
	return &v1.Result{
		Ret: v1.Retcode_INTERNAL_ERROR,
		Msg: msg,
	}
}
