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

import (
	"fmt"
	"strings"

	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	"github.com/tkeel-io/tkeel/pkg/model"
)

func EncodePluginRoute(pluginID, endpoint string) string {
	return fmt.Sprintf("%s/%s", pluginID, endpoint)
}

func DecodePluginRoute(path string) (pluginID, endpoint string) {
	sub := strings.SplitN(path, "/", 2)
	if len(sub) != 2 {
		return "", ""
	}
	sub1 := strings.Split(sub[1], "?")
	return sub[0], sub1[0]
}

func UpdatePluginRoute(sourcePluginID string, addons []*openapi_v1.ImplementedAddons, pluginRoute *model.PluginRoute) {
	if pluginRoute.RegisterAddons == nil {
		pluginRoute.RegisterAddons = make(map[string]string)
	}
	for _, v := range addons {
		pluginRoute.RegisterAddons[v.AddonsPoint] = EncodePluginRoute(sourcePluginID, v.ImplementedEndpoint)
	}
}
