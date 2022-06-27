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

package plgprofile

import (
	"math"
	"sync"

	"github.com/tkeel-io/tkeel/pkg/model"
)

const (
	// nolint
	PLUGIN_ID_KEEL = "keel"
	// api limit.
	// nolint
	MAX_API_REQUEST_LIMIT_KEY = "keel_api_request_limit"
	// nolint
	MAX_API_REQUEST_LIMIT_TITLE = "接口请求次数最大限制"
	// nolint
	DEFAULT_MAX_API_LIMIT = math.MaxInt32
)

var (
	tenantAPICount = sync.Map{}
	tenantAPILimit = sync.Map{}
)

var KeelProfiles = map[string]*model.ProfileSchema{
	MAX_API_REQUEST_LIMIT_KEY: {Type: "number", Title: MAX_API_REQUEST_LIMIT_TITLE, Description: "api请求最大次数,0 表示无限制", Default: 0, MultipleOf: 1, Maximum: DEFAULT_MAX_API_LIMIT, Minimum: 0},
}

func OnTenantAPIRequest(tenantID string, store ProfileOperator) int {
	cur, _ := tenantAPICount.Load(tenantID)
	if cur == nil {
		cur = 0
	}
	var curInt int
	if v, ok := cur.(int); ok {
		curInt = v + 1
	}
	tenantAPICount.Store(tenantID, curInt)
	return curInt
}

func GetTenantAPIRequest(tenantID string) int {
	count, ok := tenantAPICount.Load(tenantID)
	if ok {
		if value, ok := count.(int); ok {
			return value
		}
	}
	return 0
}

func SetTenantAPILimit(tenantID string, limit int) {
	tenantAPILimit.Store(tenantID, limit)
}

func ISExceededAPILimit(tenantID string) bool {
	if limited, ok := tenantAPILimit.Load(tenantID); ok {
		count, _ := tenantAPICount.Load(tenantID)
		if countVal, ok := count.(int); ok {
			if limitedVal, ok := limited.(int); ok {
				return countVal > limitedVal
			}
		}
	}
	return false
}
