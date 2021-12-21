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
package v1

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt"
)

func getSubPath(path, rootPath string) string {
	return strings.TrimPrefix(path, rootPath)
}

func getPluginIDFromApisPath(pluginPath string) string {
	ss := strings.SplitN(pluginPath, "/", 3)
	if len(ss) != 3 {
		return ""
	}
	return ss[1]
}

func getPluginMethodApisPath(pluginPath string) string {
	ss := strings.SplitN(pluginPath, "/", 3)
	if len(ss) != 3 {
		return ""
	}
	return ss[2]
}

func checkToken(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("error token invaild: %s", token)
	}
	ss := strings.Split(token, " ")
	if len(ss) != 2 {
		return "", fmt.Errorf("error token invaild: %s", token)
	}
	payload := ss[1]
	b, err := jwt.DecodeSegment(payload)
	if err != nil {
		return "", fmt.Errorf("error jwt decode: %w", err)
	}
	pmap := make(map[string]interface{})
	err = json.Unmarshal(b, &pmap)
	if err != nil {
		return "", fmt.Errorf("error json unmarshal: %w", err)
	}
	pID, ok := pmap["plugin_id"]
	if !ok {
		return "", fmt.Errorf("error token(%s) not has field: plugin_id", string(b))
	}
	pIDStr, ok := pID.(string)
	if !ok {
		return "", fmt.Errorf("error token(%s) type invaild", string(b))
	}
	return pIDStr, nil
}
