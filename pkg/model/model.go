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
	"github.com/tkeel-io/tkeel/pkg/repository"
)

type Secret struct {
	Data string `json:"data,omitempty"` // data.
}

type Installer struct {
	Repo    string `json:"repo,omitempty"`    // repo name.
	Name    string `json:"name,omitempty"`    // installer name.
	Version string `json:"version,omitempty"` // installer version.
}

type Plugin struct {
	ID                string                          `json:"id,omitempty"`                 // plugin id.
	Installer         *Installer                      `json:"installer,omitempty"`          // plugin installer.
	PluginVersion     string                          `json:"plugin_version,omitempty"`     // plugin version.
	TkeelVersion      string                          `json:"tkeel_version,omitempty"`      // plugin depend tkeel version.
	AddonsPoint       []*openapi_v1.AddonsPoint       `json:"addons_point,omitempty"`       // plugin declares addons.
	ImplementedPlugin []*openapi_v1.ImplementedPlugin `json:"implemented_plugin,omitempty"` // plugin implemented plugin list.
	Secret            *Secret                         `json:"secret,omitempty"`             // plugin registered secret.
	RegisterTimestamp int64                           `json:"register_timestamp,omitempty"` // register timestamp.
	ActiveTenantes    []string                        `json:"active_tenantes,omitempty"`    // active tenant id list.
	Version           string                          `json:"version,omitempty"`            // model version.
	State             openapi_v1.PluginStatus         `json:"state,omitempty"`              // plugin state.
}

func (p *Plugin) String() string {
	b, err := json.Marshal(p)
	if err != nil {
		return "<" + err.Error() + ">"
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
		Secret: &Secret{
			Data: p.Secret.Data,
		},
		RegisterTimestamp: p.RegisterTimestamp,
		ActiveTenantes: func() []string {
			ret := make([]string, 0, len(p.ActiveTenantes))
			copy(ret, p.ActiveTenantes)
			return ret
		}(),
		Version: p.Version,
	}
}

type PluginProxyRouteMap map[string]*PluginRoute

func (pprm *PluginProxyRouteMap) String() string {
	b, err := json.Marshal(pprm)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
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
		return "<" + err.Error() + ">"
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

func NewPlugin(pluginID string, installer *Installer) *Plugin {
	return &Plugin{
		ID:        pluginID,
		Installer: installer,
		Version:   "1",
		State:     openapi_v1.PluginStatus_UNREGISTER,
	}
}

func (p *Plugin) Register(resp *openapi_v1.IdentifyResponse, secret *Secret) {
	p.PluginVersion = resp.Version
	p.TkeelVersion = resp.TkeelVersion
	p.AddonsPoint = resp.AddonsPoint
	p.ImplementedPlugin = resp.ImplementedPlugin
	p.Secret = secret
	p.RegisterTimestamp = time.Now().Unix()
}

func NewPluginRoute(resp *openapi_v1.IdentifyResponse) *PluginRoute {
	return &PluginRoute{
		ID:           resp.PluginId,
		Status:       openapi_v1.PluginStatus_UNREGISTER,
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

type PluginRepoMap map[string]*PluginRepo

func (pprm *PluginRepoMap) String() string {
	b, err := json.Marshal(pprm)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

type Annotations map[string]interface{}

type PluginRepo struct {
	*repository.Info `json:",inline"`
	UpsertTimestamp  int64  `json:"upsert_timestamp,omitempty"` // last upsert time stamp.
	Version          string `json:"version,omitempty"`          // model version.
}

func NewPluginRepo(i *repository.Info) *PluginRepo {
	return &PluginRepo{
		Info:            i,
		UpsertTimestamp: time.Now().Unix(),
		Version:         "1",
	}
}

func (pr *PluginRepo) String() string {
	b, err := json.Marshal(pr)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
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
