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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

const (
	TKeelUser   = "_tKeel"
	TKeelTenant = "_tKeel_system"

	AdminRole = "admin"

	KeyAdminPassword = "Admin_Passwd"
)

var (
	XPluginJwtHeader = http.CanonicalHeaderKey("x-plugin-jwt")
	XtKeelAuthHeader = http.CanonicalHeaderKey("x-tKeel-auth")

	AuthorizationHeader = http.CanonicalHeaderKey("Authorization")

	TKeelComponents = []string{"rudder", "core", "keel", "security"}
)

type EnableTenant struct {
	TenantID        string `json:"tenant_id"`        // enable tenant id.
	OperatorID      string `json:"operator_id"`      // operator id.
	EnableTimestamp int64  `json:"enable_timestamp"` // enable timestamp.
}

func (et *EnableTenant) String() string {
	b, err := json.Marshal(et)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

type Plugin struct {
	ID                string                          `json:"id,omitempty"`                 // plugin id.
	Installer         *Installer                      `json:"installer,omitempty"`          // plugin installer.
	PluginVersion     string                          `json:"plugin_version,omitempty"`     // plugin version.
	TkeelVersion      string                          `json:"tkeel_version,omitempty"`      // plugin depend tkeel version.
	AddonsPoint       []*openapi_v1.AddonsPoint       `json:"addons_point,omitempty"`       // plugin declares addons.
	ImplementedPlugin []*openapi_v1.ImplementedPlugin `json:"implemented_plugin,omitempty"` // plugin implemented plugin list.
	ConsoleEntries    []*openapi_v1.ConsoleEntry      `json:"console_entries,omitempty"`    // plugin console entries.
	Secret            string                          `json:"secret,omitempty"`             // plugin registered secret.
	RegisterTimestamp int64                           `json:"register_timestamp,omitempty"` // register timestamp.
	Version           string                          `json:"version,omitempty"`            // model version.
	Status            openapi_v1.PluginStatus         `json:"status,omitempty"`             // plugin state.
	EnableTenantes    []*EnableTenant                 `json:"enable_tenantes,omitempty"`    // plugin active tenantes.
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
				ret = append(ret, &openapi_v1.AddonsPoint{
					Name: v.Name,
					Desc: v.Desc,
				})
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
							ret = append(ret, &openapi_v1.ImplementedAddons{
								AddonsPoint:         v1.AddonsPoint,
								ImplementedEndpoint: v1.ImplementedEndpoint,
							})
						}
						return ret
					}(),
				}
				ret = append(ret, tmp)
			}
			return ret
		}(),
		ConsoleEntries: func() []*openapi_v1.ConsoleEntry {
			ret := make([]*openapi_v1.ConsoleEntry, 0, len(p.ConsoleEntries))
			for _, v := range p.ConsoleEntries {
				n := &openapi_v1.ConsoleEntry{}
				consoleEntryClone(n, v)
				ret = append(ret, n)
			}
			return ret
		}(),
		Secret:            p.Secret,
		RegisterTimestamp: p.RegisterTimestamp,
		Version:           p.Version,
		Status:            p.Status,
		EnableTenantes: func() []*EnableTenant {
			ret := make([]*EnableTenant, 0, len(p.EnableTenantes))
			for _, v := range p.EnableTenantes {
				ret = append(ret, &EnableTenant{
					TenantID:        v.TenantID,
					OperatorID:      v.OperatorID,
					EnableTimestamp: v.EnableTimestamp,
				})
			}
			return ret
		}(),
	}
}

func consoleEntryClone(dst, src *openapi_v1.ConsoleEntry) {
	dst.Id = src.Id
	dst.Name = src.Name
	dst.Path = src.Path
	dst.Icon = src.Icon
	dst.Children = func() []*openapi_v1.ConsoleEntry {
		ret := make([]*openapi_v1.ConsoleEntry, 0, len(src.Children))
		for _, v := range src.Children {
			n := &openapi_v1.ConsoleEntry{}
			consoleEntryClone(n, v)
			ret = append(ret, n)
		}
		return ret
	}()
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
		Status:    openapi_v1.PluginStatus_WAIT_RUNNING,
		EnableTenantes: []*EnableTenant{
			{
				TenantID:        TKeelTenant,
				OperatorID:      TKeelUser,
				EnableTimestamp: time.Now().Unix(),
			},
		},
	}
}

func (p *Plugin) Register(resp *openapi_v1.IdentifyResponse, secret string) {
	p.PluginVersion = resp.Version
	p.TkeelVersion = resp.TkeelVersion
	p.AddonsPoint = resp.AddonsPoint
	p.ImplementedPlugin = resp.ImplementedPlugin
	p.ConsoleEntries = resp.Entries
	p.Secret = secret
	p.RegisterTimestamp = time.Now().Unix()
}

func NewPluginRoute(resp *openapi_v1.IdentifyResponse) *PluginRoute {
	return &PluginRoute{
		ID:           resp.PluginId,
		Status:       openapi_v1.PluginStatus_RUNNING,
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

func Base64Decode(p string) string {
	d, err := base64.StdEncoding.DecodeString(p)
	if err != nil {
		return err.Error()
	}
	return string(d)
}

func Base64Encode(p string) string {
	return base64.StdEncoding.EncodeToString([]byte(p))
}

type User struct {
	User   string `form:"user"`
	Tenant string `form:"tenant"`
	Role   string `form:"role"`
}

func (u *User) Base64Encode() string {
	vars := url.Values{
		"user":   []string{u.User},
		"tenant": []string{u.Tenant},
		"role":   []string{u.Role},
	}
	return base64.StdEncoding.EncodeToString([]byte(vars.Encode()))
}

func (u *User) Base64Decode(s string) error {
	d, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return fmt.Errorf("error decode(%s): %w", s, err)
	}
	vars, err := url.ParseQuery(string(d))
	if err != nil {
		return fmt.Errorf("error parse(%s): %w", string(d), err)
	}
	us, ok := vars["user"]
	if !ok {
		u.User = ""
	}
	if len(us) == 0 {
		u.User = ""
	}
	u.User = us[0]
	ts, ok := vars["tenant"]
	if !ok {
		u.Tenant = ""
	}
	if len(ts) == 0 {
		u.Tenant = ""
	}
	u.Tenant = ts[0]
	rs, ok := vars["role"]
	if !ok {
		u.Role = ""
	}
	if len(rs) == 0 {
		u.Role = ""
	}
	u.Role = rs[0]
	return nil
}
