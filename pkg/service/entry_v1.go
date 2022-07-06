package service

import (
	"context"
	"encoding/json"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/tkeel-io/kit/log"
	transport_http "github.com/tkeel-io/kit/transport/http"
	"github.com/tkeel-io/security/authz/rbac"
	v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	pb "github.com/tkeel-io/tkeel/api/entry/v1"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EntryService struct {
	pb.UnimplementedEntryServer
	pOp         plugin.Operator
	tpOp        rbac.TenantPluginMgr
	rbacOp      *casbin.SyncedEnforcer
	daprHTTPCli dapr.Client
}
type IdentifyEntries struct {
	Entries []Entry `json:"entries"`
}
type Entry struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Notifications []Notification `json:"notifications"`
}
type Notification struct {
	APIPath string `json:"api_path"`
}

type NotificationsItem struct {
	Entry        Entry       `json:"entry"`
	Notification interface{} `json:"notification"`
}

func NewEntryService(pOp plugin.Operator, tpOp rbac.TenantPluginMgr, rbacOp *casbin.SyncedEnforcer, daprHTTPCli dapr.Client) *EntryService {
	return &EntryService{
		pOp:         pOp,
		tpOp:        tpOp,
		rbacOp:      rbacOp,
		daprHTTPCli: daprHTTPCli,
	}
}

func (s *EntryService) GetEntries(ctx context.Context, req *emptypb.Empty) (*pb.GetEntriesResponse, error) {
	header := transport_http.HeaderFromContext(ctx)
	auths, ok := header[model.XtKeelAuthHeader]
	if !ok {
		log.Error("error get auth")
		return nil, pb.EntryErrInvalidTenant()
	}
	u := new(model.User)
	if err := u.Base64Decode(auths[0]); err != nil {
		log.Errorf("error decode auth(%s): %s", auths[0], err)
		return nil, pb.EntryErrInvalidTenant()
	}
	portal := v1.ConsolePortal_admin
	if u.User != model.TKeelUser {
		portal = v1.ConsolePortal_tenant
	}
	ret := make([]*v1.ConsoleEntry, 0)
	for _, v := range s.tpOp.ListTenantPlugins(u.Tenant) {
		allow, err := s.rbacOp.Enforce(u.User, u.Tenant, v, model.AllowedPermissionAction)
		if err != nil {
			log.Errorf("error rbac enforce(%s/%s/%s/%s): %s",
				u.User, u.Tenant, v, model.AllowedPermissionAction, v, err)
			return nil, pb.EntryErrInternalError()
		}
		if allow {
			if pluginIsTkeelComponent(v) {
				continue
			}
			p, err := s.pOp.Get(ctx, v)
			if err != nil {
				log.Errorf("error get plugin(%s): %s", v, err)
				if !errors.Is(err, plugin.ErrPluginNotExsist) {
					return nil, pb.EntryErrInternalError()
				}
				continue
			}
			ret = appendEntries(ret, p.ConsoleEntries, portal)
		}
	}
	sort.Sort(entrySort(ret))
	return &pb.GetEntriesResponse{
		Entries: ret,
	}, nil
}

func (s *EntryService) GetNotification(ctx context.Context, tenantID string) (interface{}, error) {
	plugins := s.tpOp.ListTenantPlugins(tenantID)
	notifications := make([]interface{}, 0)
	wg := sync.WaitGroup{}
	// 1. call console  identify & merge entry  api_path.
	for _, v := range plugins {
		if pluginIsConsole(v) {
			wg.Add(1)
			go func(pluginID string) {
				res, _ := s.daprHTTPCli.Call(ctx, &dapr.AppRequest{
					ID:     pluginID,
					Method: "v1/identify",
					Verb:   "GET",
				})
				resBytes, err := io.ReadAll(res.Body)
				if err != nil {
					log.Error(err)
					wg.Done()
					return
				}
				defer res.Body.Close()
				idfEntries := IdentifyEntries{}
				err = json.Unmarshal(resBytes, &idfEntries)
				if err != nil {
					log.Error(err)
					wg.Done()
					return
				}
				if idfEntries.Entries != nil {
					for _, entry := range idfEntries.Entries {
						if entry.Notifications != nil {
							for _, notify := range entry.Notifications {
								wg.Add(1)
								go func(apiPath string) {
									pID, route := pluginWithAPIPath(apiPath)
									notifyres, nerr := s.daprHTTPCli.Call(ctx, &dapr.AppRequest{
										ID:     pID,
										Method: route,
										Verb:   "GET",
									})
									if nerr != nil {
										log.Error(err)
										wg.Done()
										return
									}
									resNotifyBytes, _ := io.ReadAll(notifyres.Body)
									defer notifyres.Body.Close()
									resM := make(map[string]interface{})
									json.Unmarshal(resNotifyBytes, &resM)
									notificationsItem := NotificationsItem{Notification: resM["data"], Entry: entry}
									notifications = append(notifications, notificationsItem)
									wg.Done()
								}(notify.APIPath)
							}
						}
					}
				}
				wg.Done()
			}(v)
		}
	}
	wg.Wait()
	return notifications, nil
}

func appendEntries(dst, src []*v1.ConsoleEntry, portal v1.ConsolePortal) []*v1.ConsoleEntry {
	for _, v := range src {
		aP, tP := separateEntry(v)
		addEntry := aP
		if portal == v1.ConsolePortal_tenant {
			addEntry = tP
		}
		if addEntry == nil {
			continue
		}
		needMerge := false
		for i, v := range dst {
			if v.Id == addEntry.Id {
				mergeEntry(dst[i], addEntry)
				needMerge = true
				break
			}
		}
		if !needMerge {
			dst = append(dst, addEntry)
		}
	}
	sort.Sort(entrySort(dst))
	return dst
}

func separateEntry(e *v1.ConsoleEntry) (*v1.ConsoleEntry, *v1.ConsoleEntry) {
	cloneOne := proto.Clone(e)
	aP, tP := &v1.ConsoleEntry{}, &v1.ConsoleEntry{}
	proto.Merge(aP, cloneOne)
	proto.Merge(tP, cloneOne)
	if e.Children == nil {
		if e.Portal == v1.ConsolePortal_admin {
			return aP, nil
		}
		return nil, tP
	}
	aP.Children = aP.Children[0:0]
	tP.Children = tP.Children[0:0]
	for _, v := range e.Children {
		a, t := separateEntry(v)
		if a != nil {
			aP.Children = append(aP.Children, a)
		}
		if t != nil {
			tP.Children = append(tP.Children, t)
		}
	}
	if aP.Entry == "" && len(aP.Children) == 0 {
		aP = nil
	}
	if tP.Entry == "" && len(tP.Children) == 0 {
		tP = nil
	}
	return aP, tP
}

func mergeEntry(dst, src *v1.ConsoleEntry) {
	if src.Children == nil {
		return
	}
	addEntries := make([]*v1.ConsoleEntry, 0, len(src.Children))
	for _, vs := range src.Children {
		needMerge := false
		for _, vd := range dst.Children {
			if vd.Id == vs.Id {
				needMerge = true
				mergeEntry(vd, vs)
				break
			}
		}
		if !needMerge {
			addEntries = append(addEntries, vs)
		}
	}
	dst.Children = append(dst.Children, addEntries...)
	sort.Sort(entrySort(dst.Children))
}

type entrySort []*v1.ConsoleEntry

func (a entrySort) Len() int           { return len(a) }
func (a entrySort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a entrySort) Less(i, j int) bool { return a[i].Id < a[j].Id }

func pluginWithAPIPath(apiPath string) (plugin, route string) {
	s := strings.SplitN(apiPath, "/", 3)
	if len(s) == 3 {
		return s[1], s[2]
	}
	return
}
