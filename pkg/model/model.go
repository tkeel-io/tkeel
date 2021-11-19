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

package model

import (
	"encoding/json"
	"fmt"
	"time"

	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
)

type Plugin struct {
	ID                string                          `json:"id,omitempty"`                 // plugin id.
	PluginVersion     string                          `json:"plugin_version,omitempty"`     // plugin version.
	TkeelVersion      string                          `json:"tkeel_version,omitempty"`      // plugin depend tkeel version.
	AddonsPoint       []*openapi_v1.AddonsPoint       `json:"addons_point,omitempty"`       // plugin declares addons.
	ImplementedPlugin []*openapi_v1.ImplementedPlugin `json:"implemented_plugin,omitempty"` // plugin implemented plugin list.
	Secret            string                          `json:"secret,omitempty"`             // plugin registered secret.
	RegisterTimestamp int64                           `json:"register_timestamp,omitempty"` // register timestamp.
	ActiveTenantes    []string                        `json:"active_tenantes,omitempty"`    // active tenant id list.
	Version           string                          `json:"version,omitempty"`            // model version.
}

func (p *Plugin) String() string {
	b, err := json.Marshal(p)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (p *Plugin) Clone() *Plugin {
	return &Plugin{
		ID:            p.ID,
		PluginVersion: p.PluginVersion,
		TkeelVersion:  p.TkeelVersion,
		AddonsPoint: func() []*openapi_v1.AddonsPoint {
			ret := make([]*openapi_v1.AddonsPoint, 0, len(p.AddonsPoint))
			for _, v := range p.AddonsPoint {
				tmp := &openapi_v1.AddonsPoint{
					Name: v.Name,
					Desc: v.Desc,
				}
				ret = append(ret, tmp)
			}
			return ret
		}(),
		ImplementedPlugin: func() []*openapi_v1.ImplementedPlugin {
			ret := make([]*openapi_v1.ImplementedPlugin, 0, len(p.ImplementedPlugin))
			for _, v := range p.ImplementedPlugin {
				tmp := &openapi_v1.ImplementedPlugin{
					Plugin: &openapi_v1.BriefPluginInfo{
						Id:      v.Plugin.Id,
						Version: v.Plugin.Version,
					},
					Addons: func() []*openapi_v1.ImplementedAddons {
						ret := make([]*openapi_v1.ImplementedAddons, 0, len(v.Addons))
						for _, v1 := range v.Addons {
							tmp := &openapi_v1.ImplementedAddons{
								AddonsPoint:         v1.AddonsPoint,
								ImplementedEndpoint: v1.ImplementedEndpoint,
							}
							ret = append(ret, tmp)
						}
						return ret
					}(),
				}
				ret = append(ret, tmp)
			}
			return ret
		}(),
		Secret:            p.Secret,
		RegisterTimestamp: p.RegisterTimestamp,
		ActiveTenantes: func() []string {
			ret := make([]string, 0, len(p.ActiveTenantes))
			copy(ret, p.ActiveTenantes)
			return ret
		}(),
		Version: p.Version,
	}
}

type PluginRoute struct {
	ID                string                  `json:"id,omitempty"`                 // plugin id.
	Status            openapi_v1.PluginStatus `json:"status,omitempty"`             // plugin latest status.
	TkeelVersion      string                  `json:"tkeel_version,omitempty"`      // plugin depened tkeel version.
	RegisterAddons    map[string]string       `json:"register_addons,omitempty"`    // plugin register addons route map.
	ImplementedPlugin []string                `json:"implemented_plugin,omitempty"` // plugin implemented plugins.
	Version           string                  `json:"version,omitempty"`            // model version.
}

func (pr *PluginRoute) String() string {
	b, err := json.Marshal(pr)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (pr *PluginRoute) Clone() *PluginRoute {
	return &PluginRoute{
		ID:           pr.ID,
		Status:       pr.Status,
		TkeelVersion: pr.TkeelVersion,
		RegisterAddons: func() map[string]string {
			ret := make(map[string]string)
			for k, v := range pr.RegisterAddons {
				ret[k] = v
			}
			return ret
		}(),
		ImplementedPlugin: func() []string {
			ret := make([]string, 0, len(pr.ImplementedPlugin))
			copy(ret, pr.ImplementedPlugin)
			return ret
		}(),
		Version: pr.Version,
	}
}

func NewPlugin(resp *openapi_v1.IdentifyResponse, secret string) *Plugin {
	return &Plugin{
		ID:                resp.PluginId,
		PluginVersion:     resp.Version,
		TkeelVersion:      resp.TkeelVersion,
		AddonsPoint:       resp.AddonsPoint,
		ImplementedPlugin: resp.ImplementedPlugin,
		Secret:            secret,
		RegisterTimestamp: time.Now().Unix(),
		Version:           "1",
	}
}

func NewPluginRoute(resp *openapi_v1.IdentifyResponse) *PluginRoute {
	return &PluginRoute{
		ID:           resp.PluginId,
		Status:       openapi_v1.PluginStatus_STARTING,
		TkeelVersion: resp.TkeelVersion,
		ImplementedPlugin: func() []string {
			ret := make([]string, 0, len(resp.ImplementedPlugin))
			for _, v := range resp.ImplementedPlugin {
				ret = append(ret, v.Plugin.Id)
			}
			return ret
		}(),
		Version: "1",
	}
}

func Clone(src, dst interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("error marshal src: %w", err)
	}
	err = json.Unmarshal(b, dst)
	if err != nil {
		return fmt.Errorf("error unmarshal dst: %w", err)
	}
	return nil
}
