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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"
	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	"github.com/tkeel-io/tkeel/pkg/repository"
)

type Secret struct {
	Data string `json:"data,omitempty"` // data.
}

type Installer struct {
	Repo       string                   `json:"repo,omitempty"`       // repo name.
	Name       string                   `json:"name,omitempty"`       // installer name.
	Version    string                   `json:"version,omitempty"`    // installer version.
	Icon       string                   `json:"icon,omitempty"`       // installer icon.
	Desc       string                   `json:"desc,omitempty"`       // installer desc.
	Maintainer []*repository.Maintainer `json:"maintainer,omitempty"` // installer maintainer.
}

const (
	TKeelUser   = "_tKeel"
	TKeelTenant = "_tKeel_system"

	AdminRole = "admin"

	KeyAdminPassword = "admin_passwd"

	KeyPermissionSet = "permission_set"

	AllowedPermissionAction = "_tkeel_allow"

	_allowedPluginAccessName = " 允许访问"
	_allowedPluginAccessDesc = "访问插件的权限，当权限被允许时，插件声明的菜单将会开放"

	TkeelTenantAdminRole = "管理员"
	TKeelTenantAdminDesc = "默认拥有租户空间下所有权限"
)

var (
	ErrPermissionExist                   = errors.New("Permission exist")
	ErrPermissionNotExist                = errors.New("Permission not exist")
	ErrPermissionNotRoot                 = errors.New("Permission not be root")
	ErrDuplicatePermissionAsTheSameLevel = errors.New("Duplicate permissions at the same level")
	ErrPermissionDependencyNotExist      = errors.New("Permission dependency does not exist")
)

var (
	XPluginJwtHeader = http.CanonicalHeaderKey("x-plugin-jwt")
	XtKeelAuthHeader = http.CanonicalHeaderKey("x-tKeel-auth")

	AuthorizationHeader = http.CanonicalHeaderKey("Authorization")

	TKeelComponents = []string{"rudder", "core", "keel", "security"}

	_permissionSet = NewPermissionSet()
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
	PluginDependences []*openapi_v1.BriefPluginInfo   `json:"plugin_dependences,omitempty"` // plugin dependences.
	Permissions       []*openapi_v1.Permission        `json:"permissions,omitempty"`        // plugin permissions.
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
	p.PluginDependences = resp.Dependence
	p.ConsoleEntries = resp.Entries
	p.Permissions = resp.Permissions
	p.Secret = secret
	p.RegisterTimestamp = time.Now().Unix()
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
		PluginDependences: func() []*openapi_v1.BriefPluginInfo {
			ret := make([]*openapi_v1.BriefPluginInfo, 0, len(p.PluginDependences))
			for _, v := range p.PluginDependences {
				ret = append(ret, &openapi_v1.BriefPluginInfo{
					Id:      v.Id,
					Version: v.Version,
				})
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

func (p *Plugin) CheckTenantEnable(tenantID string) bool {
	for _, v := range p.EnableTenantes {
		if v.TenantID == tenantID {
			return true
		}
	}
	return false
}

func (p *Plugin) TenantEnable(t *EnableTenant) {
	p.EnableTenantes = append(p.EnableTenantes, t)
}

func (p *Plugin) TenantDisable(tenantID string) bool {
	ok := false
	for i, v := range p.EnableTenantes {
		if v.TenantID == tenantID {
			p.EnableTenantes = append(p.EnableTenantes[:i], p.EnableTenantes[i+1:]...)
			ok = true
			break
		}
	}
	return ok
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

func ClonePermission(pb *openapi_v1.Permission) *openapi_v1.Permission {
	clonePb := &openapi_v1.Permission{}
	proto.Merge(proto.Clone(pb), clonePb)
	return clonePb
}

type Permission struct {
	Path string
	Pb   *openapi_v1.Permission
}

type PermissionSet struct {
	rwLock   *sync.RWMutex
	rawSet   map[string]*openapi_v1.Permission
	sortList []*Permission
	pathSet  map[string]*Permission
}

func NewPermissionSet() *PermissionSet {
	return &PermissionSet{
		rwLock:   new(sync.RWMutex),
		rawSet:   make(map[string]*openapi_v1.Permission),
		sortList: make([]*Permission, 0),
		pathSet:  make(map[string]*Permission),
	}
}

func GetPermissionSet() *PermissionSet {
	return _permissionSet
}

func (ps *PermissionSet) Marshall() ([]byte, error) {
	ps.rwLock.RLock()
	defer ps.rwLock.RUnlock()
	b, err := json.Marshal(&(ps.rawSet))
	if err != nil {
		return nil, errors.Wrap(err, "marshal permission set")
	}
	return b, nil
}

func (ps *PermissionSet) Unmarshal(b []byte) error {
	ps.rwLock.Lock()
	defer ps.rwLock.Unlock()
	if err := json.Unmarshal(b, &(ps.rawSet)); err != nil {
		return errors.Wrapf(err, "unmarshal permission set(%s)", string(b))
	}
	ps.sortList = ps.sortList[0:0]
	ps.pathSet = make(map[string]*Permission)
	for pluginID, v := range ps.rawSet {
		if pluginID == v.Id {
			for _, p := range convertPB2Model(pluginID, v) {
				ps.sortList = append(ps.sortList, p)
				ps.pathSet[p.Path] = p
			}
		}
	}
	sort.Sort(PermissionSort(ps.sortList))
	return nil
}

func (ps *PermissionSet) GetSortList() []*Permission {
	ret := make([]*Permission, 0, len(ps.sortList))
	ps.rwLock.RLock()
	defer ps.rwLock.RUnlock()
	copy(ret, ps.sortList)
	return ret
}

func (ps *PermissionSet) GetPermissionByPluginID(pluginID string) []*Permission {
	ret := make([]*Permission, 0, len(ps.sortList))
	ps.rwLock.RLock()
	defer ps.rwLock.RUnlock()
	for _, v := range ps.sortList {
		if strings.HasPrefix(v.Path, pluginID) {
			ret = append(ret, v)
		}
	}
	return ret
}

func (ps *PermissionSet) NewPluginAllowedPermission(pluginID string) *Permission {
	return &Permission{
		Path: pluginID,
		Pb: &openapi_v1.Permission{
			Id:   pluginID,
			Name: pluginID + _allowedPluginAccessName,
			Desc: _allowedPluginAccessDesc,
		},
	}
}

func (ps *PermissionSet) Add(pluginID string, pb *openapi_v1.Permission) (bool, error) {
	if pluginID != pb.Id {
		return false, ErrPermissionNotRoot
	}
	ps.rwLock.RLock()
	if _, ok := (ps.rawSet)[pluginID]; ok {
		return false, ErrPermissionExist
	}
	ps.rwLock.RUnlock()
	if err := ps.checkPermission(pb); err != nil {
		return false, errors.Wrap(err, "check permission")
	}
	pList := convertPB2Model("", pb)
	ps.rwLock.Lock()
	defer ps.rwLock.Unlock()
	(ps.rawSet)[pluginID] = pb
	for _, v := range pList {
		ps.pathSet[v.Path] = v
		ps.sortList = append(ps.sortList, v)
	}
	sort.Sort(PermissionSort(ps.sortList))
	return true, nil
}

func (ps *PermissionSet) Delete(pluginID string) {
	ps.rwLock.Lock()
	defer ps.rwLock.Unlock()
	delete((ps.rawSet), pluginID)
	for k := range ps.pathSet {
		if strings.HasPrefix(k, pluginID) {
			delete(ps.pathSet, k)
		}
	}
	tmp := make([]*Permission, 0, len(ps.sortList))
	for _, v := range ps.sortList {
		if !strings.HasPrefix(v.Path, pluginID) {
			tmp = append(tmp, v)
		}
	}
	ps.sortList = tmp
}

func (ps *PermissionSet) checkPermission(p *openapi_v1.Permission) error {
	for _, v := range p.Dependences {
		if _, err := ps.GetPermission(v.Path); err != nil {
			if errors.Is(err, ErrPermissionNotExist) {
				return ErrPermissionDependencyNotExist
			}
		}
	}
	idSet := make(map[string]struct{}, len(p.Children))
	for _, v := range p.Children {
		if _, ok := idSet[v.Id]; ok {
			return ErrDuplicatePermissionAsTheSameLevel
		}
		idSet[v.Id] = struct{}{}
		if err := ps.checkPermission(v); err != nil {
			return errors.Wrapf(err, "check child permission(%s)", v)
		}
	}
	return nil
}

func (ps *PermissionSet) GetPermission(path string) (*Permission, error) {
	ps.rwLock.RLock()
	defer ps.rwLock.RUnlock()
	if path == "" {
		return nil, ErrPermissionNotExist
	}
	p, ok := ps.pathSet[path]
	if !ok {
		return nil, ErrPermissionNotExist
	}
	return p, nil
}

func convertPB2Model(parentalPath string, pb *openapi_v1.Permission) []*Permission {
	ret := make([]*Permission, 0)
	path := parentalPath + "/" + pb.Id
	if parentalPath == "" || parentalPath == pb.Id {
		path = pb.Id
	}
	ret = append(ret, &Permission{
		Path: path,
		Pb:   pb,
	})
	for _, v := range pb.Children {
		ret = append(ret, convertPB2Model(path, v)...)
	}
	return ret
}

type PermissionSort []*Permission

func (a PermissionSort) Len() int      { return len(a) }
func (a PermissionSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a PermissionSort) Less(i, j int) bool {
	si := strings.Split(a[i].Path, "/")
	sj := strings.Split(a[j].Path, "/")
	lenth := len(si)
	if len(si) > len(sj) {
		lenth = len(sj)
	}
	for index := 0; index < lenth; index++ {
		if si[index] != sj[index] {
			return si[index] < sj[index]
		}
	}
	return len(si) < len(sj)
}

type Role struct {
	TenantID       string   `json:"tenant_id"`
	Name           string   `json:"name"`
	Desc           string   `json:"desc"`
	PermissionPath []string `json:"permission_path"`
}
