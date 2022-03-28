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
	"strconv"
	"sync"

	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	pb "github.com/tkeel-io/tkeel/api/profile/v1"
)

const (
	PLUGIN_ID_KEEL = "keel"
	// api limit.
	MAX_API_REQUEST_LIMIT_KEY  = "max_api_request_limit"
	MAX_API_REQUEST_LIMIT_DESC = "接口请求次数最大限制"
	DEFAULT_MAX_API_LIMIT      = math.MaxInt
)

var (
	tenantAPICount = sync.Map{}
	tenantAPILimit = sync.Map{}
)

var KeelProfiles = &pb.TenantProfiles{PluginId: PLUGIN_ID_KEEL, Profiles: []*openapi_v1.ProfileItem{{Key: MAX_API_REQUEST_LIMIT_KEY,
	Val: strconv.Itoa(DEFAULT_MAX_API_LIMIT), Description: MAX_API_REQUEST_LIMIT_DESC, ValType: openapi_v1.TypeProfileVal_NUMBER},
}}

func OnTenantAPIRequest(tenantID string) int {
	count, ok := tenantAPICount.Load(tenantID)
	if ok {
		countInt := count.(int) + 1
		tenantAPICount.Store(tenantID, countInt)
		return countInt
	}
	tenantAPICount.Store(tenantID, 1)
	return 1
}

func GetTenantAPIRequest(tenantID string) int {
	count, ok := tenantAPICount.Load(tenantID)
	if ok {
		return count.(int)
	}
	return 0
}

func SetTenantAPILimit(tenantID string, limit int) {
	tenantAPILimit.Store(tenantID, limit)
}

func ISExceededAPILimit(tenantID string) bool {
	limited, ok := tenantAPILimit.Load(tenantID)
	if ok {
		count, _ := tenantAPICount.Load(tenantID)
		return count.(int) > limited.(int)
	}
	return false
}
