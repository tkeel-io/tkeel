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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/security/authz/rbac"
	s_model "github.com/tkeel-io/security/model"
	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	pb "github.com/tkeel-io/tkeel/api/plugin/v1"
	"github.com/tkeel-io/tkeel/pkg/client/openapi"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/hub"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/kv"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	"github.com/tkeel-io/tkeel/pkg/repository"
	"github.com/tkeel-io/tkeel/pkg/repository/helm"
	"github.com/tkeel-io/tkeel/pkg/util"
	"github.com/tkeel-io/tkeel/pkg/version"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

var (
	ErrGetOpenapiIdentify = errors.New("error get openapi identify")
	ErrPluginRegistered   = errors.New("plugin is registered")
)

type PluginServiceV1 struct {
	pb.UnimplementedPluginServer

	tkeelConf      *config.TkeelConf
	kvOp           kv.Operator
	pluginOp       plugin.Operator
	pluginRouteOp  proute.Operator
	tenantPluginOp rbac.TenantPluginMgr
	openapiClient  openapi.Client
	db             *gorm.DB
}

func NewPluginServiceV1(db *gorm.DB, conf *config.TkeelConf, kvOp kv.Operator, pOp plugin.Operator,
	prOp proute.Operator, tpOp rbac.TenantPluginMgr, openapi openapi.Client) *PluginServiceV1 {
	ok, err := tpOp.OnCreateTenant(model.TKeelTenant)
	if err != nil {
		log.Fatalf("error create tenant %s: %s", model.TKeelTenant, err)
	}
	log.Debugf("tKeel create tenant %s %v", model.TKeelTenant, ok)
	for _, v := range model.TKeelComponents {
		ok, err = tpOp.AddTenantPlugin(model.TKeelTenant, v)
		if err != nil {
			log.Fatalf("error %s enable %s: %s", model.TKeelTenant, v, err)
		}
		log.Debugf("tKeel enable %s %v", v, ok)
	}
	return &PluginServiceV1{
		tkeelConf:      conf,
		kvOp:           kvOp,
		pluginOp:       pOp,
		pluginRouteOp:  prOp,
		tenantPluginOp: tpOp,
		openapiClient:  openapi,
		db:             db,
	}
}

func (s *PluginServiceV1) InstallPlugin(ctx context.Context,
	req *pb.InstallPluginRequest) (*pb.InstallPluginResponse, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	if req.Installer == nil {
		log.Error("error install plugin request installer info is nil")
		return nil, pb.PluginErrInvalidArgument()
	}
	installerConfiguration, err := getInstallerConfiguration(req)
	if err != nil {
		log.Errorf("error get installer configuration: %s", err)
		return nil, pb.PluginErrInvalidArgument()
	}
	log.Debugf("configuration: %v", installerConfiguration)
	repo, err := hub.GetInstance().Get(req.Installer.Repo)
	if err != nil {
		log.Errorf("error get repo(%s): %s", req.Installer.Repo, err)
		if errors.Is(err, hub.ErrRepoNotFound) {
			return nil, pb.PluginErrInstallerNotFound()
		}
	}
	installer, err := repo.Get(req.Installer.Name, req.Installer.Version)
	if err != nil {
		log.Errorf("error get installer(%s): %s", req.Installer, err)
		return nil, pb.PluginErrInstallerNotFound()
	}
	installer.SetPluginID(req.Id)
	if err = installer.Install(convertConfiguration2Option(installerConfiguration)...); err != nil {
		log.Errorf("error install installer(%s) err: %s", installer.Brief(), err)
		if errors.Is(err, repository.ErrInvalidOptions) {
			return nil, pb.PluginErrInvalidArgument()
		}
		return nil, pb.PluginErrInstallInstaller()
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("installer roll back.")
		if err = hub.GetInstance().Uninstall(req.Id, installer.Brief()); err != nil {
			return fmt.Errorf("error uninstall installer(%s): %w",
				installer.Brief(), err)
		}
		return nil
	})
	// create new plugin.
	newP := model.NewPlugin(req.Id, &model.Installer{
		Repo:    installer.Brief().Repo,
		Name:    installer.Brief().Name,
		Version: installer.Brief().Version,
		Icon:    installer.Brief().Icon,
		Desc:    installer.Brief().Desc,
	})
	if err = s.pluginOp.Create(ctx, newP); err != nil {
		log.Errorf("error create plugin(%s): %s", newP, err)
		if errors.Is(err, plugin.ErrPluginExsist) {
			return nil, pb.PluginErrPluginAlreadyExists()
		}
		return nil, pb.PluginErrInternalStore()
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("installer roll back.")
		if _, err = s.pluginOp.Delete(ctx, req.Id); err != nil {
			return fmt.Errorf("error delete plugin(%s): %w", req.Id, err)
		}
		return nil
	})
	// tkeel tenant enalbe plugin.
	if _, err = s.tenantPluginOp.AddTenantPlugin(model.TKeelTenant, req.Id); err != nil {
		log.Errorf("error add tenant(%s) plugin(%s): %s", model.TKeelTenant, req.Id, err)
		return nil, pb.PluginErrUnknown()
	}
	rbStack = util.NewRollbackStack()
	go func() {
		actionCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
		defer cancel()
		s.registerPluginAction(actionCtx, newP.ID)
	}()
	log.Debugf("install plugin(%s) succ.", newP)
	return &pb.InstallPluginResponse{
		Plugin: util.ConvertModel2PluginObjectPb(newP, nil, model.TKeelTenant),
	}, nil
}

func (s *PluginServiceV1) UninstallPlugin(ctx context.Context,
	req *pb.UninstallPluginRequest) (*pb.UninstallPluginResponse, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	p, err := s.pluginOp.Get(ctx, req.GetId())
	if err != nil {
		log.Errorf("error plugin operator get: %s", err)
		if errors.Is(err, plugin.ErrPluginNotExsist) {
			return nil, pb.PluginErrPluginNotFound()
		}
		return nil, pb.PluginErrInternalStore()
	}
	pr, err := s.pluginRouteOp.Get(ctx, req.GetId())
	if err != nil {
		if !errors.Is(err, proute.ErrPluginRouteNotExsist) {
			log.Errorf("error plugin operator get: %s", err)
			return nil, pb.PluginErrInternalStore()
		}
	}
	if p.Installer == nil {
		log.Errorf("error plugin(%s) installer is nil", p)
		return nil, pb.PluginErrInternalStore()
	}
	// check whether the extension point is implemented.
	if pr != nil && len(pr.RegisterAddons) != 0 {
		log.Errorf("error uninstall plugin(%s): other plugins have implemented the extension points of this plugin.", req.GetId())
		return nil, pb.PluginErrUninstallPluginHasBeenDepended()
	}
	// Check if plugin is disabled by tenant.
	if len(p.EnableTenantes) > 1 {
		log.Errorf("error unregister plugin(%s): tenant(%v) enableed", p.ID, p.EnableTenantes)
		return nil, pb.PluginErrPluginHasTenantEnabled()
	}
	// reset implemented plugin route.
	subRbStack, err := s.resetImplementedPluginRoute(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("error reset implemented plugin route(%s): %w", p.ID, err)
	}
	rbStack = append(rbStack, subRbStack...)
	// delete plugin.
	rb, err := s.deletePlugin(ctx, req.GetId())
	if err != nil {
		log.Errorf("error delete plugin(%s): %s", req.GetId(), err)
		return nil, pb.PluginErrUninstallPlugin()
	}
	rbStack = append(rbStack, rb)
	// delete plugin route.
	rb, err = s.deletePluginRoute(ctx, req.GetId())
	if err != nil {
		log.Errorf("error delete plugin route(%s): %s", req.GetId(), err)
		return nil, pb.PluginErrUninstallPlugin()
	}
	rbStack = append(rbStack, rb)
	// tkeel tenant disable plugin.
	if _, err = s.tenantPluginOp.DeleteTenantPlugin(model.TKeelTenant, req.Id); err != nil {
		log.Errorf("error delete tenant(%s) plugin(%s): %s", model.TKeelTenant, req.Id, err)
		return nil, pb.PluginErrUnknown()
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("uninstall plugin  tenant(%s) plugin(%s) route roll back run.", model.TKeelTenant, req.Id)
		if _, err = s.tenantPluginOp.AddTenantPlugin(model.TKeelTenant, req.Id); err != nil {
			return fmt.Errorf("error add tenant(%s) plugin(%s): %w", model.TKeelTenant, req.Id, err)
		}
		return nil
	})
	// uninstall plugin.
	if err = hub.GetInstance().Uninstall(req.GetId(), &repository.InstallerBrief{
		Name:      p.Installer.Name,
		Repo:      p.Installer.Repo,
		Version:   p.Installer.Version,
		Installed: true,
	}); err != nil {
		log.Errorf("error uninstall plugin(%s): %s", p, err)
		return nil, pb.PluginErrUninstallPlugin()
	}
	log.Debugf("uninstall plugin(%s) succ.", p)
	rbStack = util.NewRollbackStack()
	return &pb.UninstallPluginResponse{
		Plugin: util.ConvertModel2PluginObjectPb(p, pr, model.TKeelTenant),
	}, nil
}

func (s *PluginServiceV1) GetPlugin(ctx context.Context,
	req *pb.GetPluginRequest) (*pb.GetPluginResponse, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user", err)
		return nil, pb.PluginErrUnknown()
	}
	gPlugin, err := s.pluginOp.Get(ctx, req.Id)
	if err != nil {
		log.Errorf("error plugin(%s) get: %s", req.Id, err)
		if errors.Is(err, plugin.ErrPluginNotExsist) {
			return nil, pb.PluginErrPluginNotFound()
		}
		return nil, pb.PluginErrInternalStore()
	}
	gPluginRoute, err := s.pluginRouteOp.Get(ctx, req.Id)
	if err != nil {
		if !errors.Is(err, proute.ErrPluginRouteNotExsist) {
			log.Errorf("error plugin(%s) route get: %s", req.Id, err)
			return nil, pb.PluginErrInternalStore()
		}
	}

	return &pb.GetPluginResponse{
		Plugin: util.ConvertModel2PluginObjectPb(gPlugin, gPluginRoute, u.Tenant),
	}, nil
}

func (s *PluginServiceV1) ListPlugin(ctx context.Context,
	req *pb.ListPluginRequest) (*pb.ListPluginResponse, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user", err)
		return nil, pb.PluginErrUnknown()
	}
	ps, err := s.pluginOp.List(ctx)
	if err != nil {
		log.Errorf("error plugin list: %s", err)
		return nil, pb.PluginErrListPlugin()
	}
	pList := make(pluginList, 0, len(ps))
	for _, p := range ps {
		pList = append(pList, util.ConvertModel2PluginBriefObjectPb(p, u.Tenant))
	}

	regular := getReglarStringKeyWords(req.KeyWords)
	exp, err := regexp.Compile(regular)
	if err != nil {
		log.Errorf("error compile regular(%s): %s", regular, err)
		return nil, pb.PluginErrInvalidArgument()
	}
	ret := make(pluginList, 0, len(pList))
	for _, v := range pList {
		if exp.MatchString(v.Id) {
			ret = append(ret, v)
		}
	}
	total := ret.Len()
	sort.Sort(ret)
	start, end := getQueryItemsStartAndEnd(int(req.PageNum), int(req.PageSize), len(ret))
	return &pb.ListPluginResponse{
		Total:      int32(total),
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		PluginList: ret[start:end],
	}, nil
}

func (s *PluginServiceV1) TenantEnable(ctx context.Context,
	req *pb.TenantEnableRequest) (*emptypb.Empty, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.PluginErrInvalidArgument()
	}
	p, err := s.pluginOp.Get(ctx, req.Id)
	if err != nil {
		log.Errorf("error get plugin route: %s", err)
		return nil, pb.PluginErrInternalStore()
	}
	for _, v := range p.EnableTenantes {
		if v.TenantID == u.Tenant {
			log.Errorf("error tenant(%s) has been enabled", v)
			return nil, pb.PluginErrDuplicateEnableTenant()
		}
	}
	// openapi tenant/enable.
	rb, err := s.requestTenantEnable(ctx, req.Id, u.Tenant, req.Extra.Extra)
	if err != nil {
		log.Errorf("error request(%s) tenant(%s/%s/%s) enable: %s",
			req.Id, u.Tenant, req.Extra.Desc, string(req.Extra.Extra), err)
		return nil, pb.PluginErrOpenapiEnabletenant()
	}
	rbStack = append(rbStack, rb)
	// add tenant plugin rbac.
	if _, err = s.tenantPluginOp.AddTenantPlugin(u.Tenant, req.Id); err != nil {
		log.Errorf("error add tenant(%s) plugin(%s) rbac: %s", u.Tenant, p, err)
		return nil, pb.PluginErrUnknown()
	}
	// update plugin.
	tmpP := p.Clone()
	p.EnableTenantes = append(p.EnableTenantes, &model.EnableTenant{
		TenantID:        u.Tenant,
		OperatorID:      u.User,
		EnableTimestamp: time.Now().Unix(),
	})
	if err = s.pluginOp.Update(ctx, p); err != nil {
		log.Errorf("error enable tenant(%s) update plugin(%s): %s", u.Tenant, p, err)
		return nil, pb.PluginErrInternalStore()
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("roll back enable tenant: %s --> %s", p, tmpP)
		if _, err = s.pluginOp.Delete(ctx, tmpP.ID); err != nil {
			log.Errorf("error roll back update plugin(%s) delete: %s", tmpP, err)
			return fmt.Errorf("error p delete: %w", err)
		}
		tmpP.Version = "1"
		if err = s.pluginOp.Create(ctx, tmpP); err != nil {
			log.Errorf("error roll back update plugin(%s) create: %s", tmpP, err)
			return fmt.Errorf("error p create: %w", err)
		}
		return nil
	})
	rbStack = util.NewRollbackStack()
	log.Debugf("tenant(%s) enable plugin(%s) succ.", u.Tenant, req.Id)
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) TenantDisable(ctx context.Context,
	req *pb.TenantDisableRequest) (*emptypb.Empty, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.PluginErrInvalidArgument()
	}
	p, err := s.pluginOp.Get(ctx, req.Id)
	if err != nil {
		log.Errorf("error get plugin route: %s", err)
		return nil, pb.PluginErrInternalStore()
	}

	update := false
	tmpP := p.Clone()
	for i, v := range p.EnableTenantes {
		if v.TenantID == u.Tenant {
			p.EnableTenantes = append(p.EnableTenantes[:i], p.EnableTenantes[i+1:]...)
			update = true
			break
		}
	}
	if update {
		// openapi tenant/disable.
		rb, err := s.requestTenantDisable(ctx, req.Id, u.Tenant, req.Extra)
		if err != nil {
			log.Errorf("error request(%s) tenant(%s/%s) disable: %s",
				req.Id, u.Tenant, string(req.Extra), err)
			return nil, pb.PluginErrOpenapiDisableTenant()
		}
		rbStack = append(rbStack, rb)
		// delete tenant plugin rbac.
		if _, err = s.tenantPluginOp.DeleteTenantPlugin(u.Tenant, req.Id); err != nil {
			log.Errorf("error delete tenant(%s) plugin(%s) rbac: %s", u.Tenant, p, err)
			return nil, pb.PluginErrUnknown()
		}
		// plugin update.
		if err = s.pluginOp.Update(ctx, p); err != nil {
			log.Errorf("error disable tenant(%s) update plugin(%s)", u.Tenant, p)
			return nil, pb.PluginErrInternalStore()
		}
		rbStack = append(rbStack, func() error {
			log.Debugf("roll back disable tenant: %s --> %s", p, tmpP)
			if _, err = s.pluginOp.Delete(ctx, tmpP.ID); err != nil {
				log.Errorf("error roll back update plugin(%s) delete: %s", tmpP, err)
				return fmt.Errorf("error p delete: %w", err)
			}
			tmpP.Version = "1"
			if err = s.pluginOp.Create(ctx, tmpP); err != nil {
				log.Errorf("error roll back update plugin(%s) create: %s", tmpP, err)
				return fmt.Errorf("error p create: %w", err)
			}
			return nil
		})
	}
	rbStack = util.NewRollbackStack()
	log.Debugf("tenant(%s) disable plugin(%s) succ.", u.Tenant, req.Id)
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) ListEnabledTenants(ctx context.Context,
	req *pb.ListEnabledTenantsRequest) (*pb.ListEnabledTenantsResponse, error) {
	p, err := s.pluginOp.Get(ctx, req.Id)
	if err != nil {
		log.Errorf("error get plugin: %s", err)
		if errors.Is(err, plugin.ErrPluginNotExsist) {
			return nil, pb.PluginErrPluginNotFound()
		}
		return nil, pb.PluginErrInternalStore()
	}

	regular := getReglarStringKeyWords(req.KeyWords)
	exp, err := regexp.Compile(regular)
	if err != nil {
		log.Errorf("error compile regular(%s): %s", regular, err)
		return nil, pb.PluginErrInvalidArgument()
	}
	daoTenant := &s_model.Tenant{}
	daoUser := &s_model.User{}
	ret := make([]*pb.EnabledTenant, 0, len(p.EnableTenantes))
	for _, v := range p.EnableTenantes {
		if v.TenantID == model.TKeelTenant {
			continue
		}
		if exp.MatchString(v.TenantID) {
			where := map[string]interface{}{"id": v.TenantID}
			_, ts, err := daoTenant.List(s.db, where, nil, "")
			if err != nil {
				log.Warnf("error list tenant(%s): %s", v.TenantID, err)
				continue
			}
			if len(ts) != 1 {
				log.Warnf("error list tenant(%s/%d) invalid", v.TenantID, len(ts))
				continue
			}
			num, err := daoUser.CountInTenant(s.db, v.TenantID)
			if err != nil {
				log.Warnf("error count user in tenant(%s): %s", v.TenantID, err)
				continue
			}
			ret = append(ret, &pb.EnabledTenant{
				TenantId:        v.TenantID,
				OperatorId:      v.OperatorID,
				EnableTimestamp: v.EnableTimestamp,
				Title:           ts[0].Title,
				Remark:          ts[0].Remark,
				UserNum:         int32(num),
			})
		}
	}
	etList := enabledTenantList(ret)
	sort.Sort(etList)
	total := etList.Len()
	start, end := getQueryItemsStartAndEnd(int(req.PageNum), int(req.PageSize), total)
	etList = etList[start:end]
	return &pb.ListEnabledTenantsResponse{
		Total:    int32(total),
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Tenants:  etList,
	}, nil
}

func (s *PluginServiceV1) registerPluginAction(ctx context.Context, pID string) {
	duration, err := time.ParseDuration(s.tkeelConf.WatchInterval)
	if err != nil {
		log.Errorf("register error parse watch interval: %s", err)
		return
	}

	ticker := time.NewTicker(duration * 10)
	registrationfailed := true
	defer func() {
		if registrationfailed {
			if err = s.updatePluginStatus(ctx, pID, openapi_v1.PluginStatus_ERR_REGISTER); err != nil {
				log.Errorf("register update register error plugin: %s", err)
			}
		}
	}()
	retry := 5
	log.Debugf("start register plugin(%s) retry: %d", pID, retry)
	for {
		select {
		case <-ticker.C:
			resp, err := s.queryStatus(ctx, pID)
			if err != nil {
				log.Warnf("register query plugin(%s) status: %s retry: %d", pID, err, retry)
				if retry == 0 {
					return
				}
				retry--
			} else if resp.Status == openapi_v1.PluginStatus_RUNNING {
				// get register plugin identify.
				resp, err := s.queryIdentify(ctx, pID)
				if err != nil {
					log.Errorf("register error query identify: %s", err)
					return
				}
				// check register plugin identify.
				if err = s.checkIdentify(ctx, resp); err != nil {
					log.Errorf("register error check identify: %s", err)
					return
				}
				if err = s.verifyPluginIdentity(ctx, resp); err != nil {
					log.Errorf("register error register plugin: %s", err)
					return
				}
				registrationfailed = false
				log.Debugf("register plugin(%s) ok", pID)
				return
			}
			ticker.Reset(duration)
		case <-ctx.Done():
			log.Errorf("register plugin(%s) timeout", pID)
			return
		}
	}
}

func (s *PluginServiceV1) updatePluginStatus(ctx context.Context, pID string, status openapi_v1.PluginStatus) error {
	// get plugin.
	p, err := s.pluginOp.Get(ctx, pID)
	if err != nil {
		return fmt.Errorf("error get plugin: %w", err)
	}
	// update plugin status.
	p.Status = status
	if err = s.pluginOp.Update(ctx, p); err != nil {
		return fmt.Errorf("error update plugin: %w", err)
	}
	return nil
}

func (s *PluginServiceV1) queryIdentify(ctx context.Context,
	pID string) (*openapi_v1.IdentifyResponse, error) {
	if pID == "" {
		return nil, errors.New("error empty plugin id")
	}
	identifyResp, err := s.openapiClient.Identify(ctx, pID)
	if err != nil {
		return nil, fmt.Errorf("error identify(%s): %w", pID, err)
	}
	if identifyResp.Res == nil {
		return nil, fmt.Errorf("error identify(%s): Res is nil", pID)
	}
	if identifyResp.Res.Ret != openapi_v1.Retcode_OK {
		return nil, fmt.Errorf("error identify(%s): %s", pID, identifyResp.Res.Ret)
	}
	if identifyResp.PluginId != pID {
		return nil, fmt.Errorf("error plugin id not match: %s -- %s", pID, identifyResp.PluginId)
	}
	return identifyResp, nil
}

func (s *PluginServiceV1) queryStatus(ctx context.Context, pID string) (*openapi_v1.StatusResponse, error) {
	if pID == "" {
		return nil, errors.New("error empty plugin id")
	}
	statusResp, err := s.openapiClient.Status(ctx, pID)
	if err != nil {
		return nil, fmt.Errorf("error status(%s): %w", pID, err)
	}
	if statusResp.Res == nil {
		return nil, fmt.Errorf("error status(%s): Res is nil", pID)
	}
	if statusResp.Res.Ret != openapi_v1.Retcode_OK {
		return nil, fmt.Errorf("error status(%s): %s", pID, statusResp.Res.Ret)
	}
	return statusResp, nil
}

func (s *PluginServiceV1) checkIdentify(ctx context.Context,
	resp *openapi_v1.IdentifyResponse) error {
	ok, err := util.CheckRegisterPluginTkeelVersion(resp.TkeelVersion, version.Version)
	if err != nil {
		return fmt.Errorf("error check register plugin(%s) depend tkeel version(%s): %w",
			resp.PluginId, resp.TkeelVersion, err)
	}
	if !ok {
		return fmt.Errorf("error plugin(%s) depend tkeel version(%s) not invalid",
			resp.PluginId, resp.TkeelVersion)
	}
	// TODO: openapi depence plugin check.
	return nil
}

func (s *PluginServiceV1) verifyPluginIdentity(ctx context.Context, resp *openapi_v1.IdentifyResponse) error {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	// create plugin route.
	newPluginRoute := model.NewPluginRoute(resp)
	err := s.pluginRouteOp.Create(ctx, newPluginRoute)
	if err != nil {
		if errors.Is(err, proute.ErrPluginRouteExsist) {
			return ErrPluginRegistered
		}
		return fmt.Errorf("error create new plugin route(%s): %w", newPluginRoute, err)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("save register plugin route roll back run.")
		_, err = s.pluginRouteOp.Delete(ctx, newPluginRoute.ID)
		if err != nil {
			return fmt.Errorf("error delete new plugin(%s) route: %w", newPluginRoute.ID, err)
		}
		return nil
	})
	// register implemented plugin route.
	rbs, err := s.checkImplementedPluginRoute(ctx, resp)
	if err != nil {
		return fmt.Errorf("error register implemented plugin(%s) route: %w", resp.PluginId, err)
	}
	rbStack = append(rbStack, rbs...)
	// update register plugin and update plugin route.
	p, err := s.pluginOp.Get(ctx, resp.PluginId)
	if err != nil {
		return fmt.Errorf("error get plugin(%s): %w", resp.PluginId, err)
	}
	p.Register(resp, helm.SecretContext)
	p.Status = openapi_v1.PluginStatus_RUNNING
	if err := s.pluginOp.Update(ctx, p); err != nil {
		return fmt.Errorf("error update plugin(%s): %w", p, err)
	}
	rbStack = util.NewRollbackStack()
	return nil
}

func (s *PluginServiceV1) checkImplementedPluginRoute(ctx context.Context,
	resp *openapi_v1.IdentifyResponse) (util.RollBackStack, error) {
	rbStack := util.NewRollbackStack()
	for _, v := range resp.ImplementedPlugin {
		oldPluginRoute, err := s.pluginRouteOp.Get(ctx, v.Plugin.Id)
		if err != nil {
			rbStack.Run()
			return nil, fmt.Errorf("error implemented plugin(%s) not registered", v.Plugin.Id)
		}
		ok, err := util.CheckRegisterPluginTkeelVersion(oldPluginRoute.TkeelVersion, resp.TkeelVersion)
		if err != nil {
			rbStack.Run()
			return nil, fmt.Errorf("error check implemented plugin(%s) depened tkeel version: %w",
				v.Plugin.Id, err)
		}
		if !ok {
			rbStack.Run()
			return nil, fmt.Errorf(`error implemented plugin(%s) depened tkeel version(%s) 
			not less register plugin tkeel version(%s)`,
				v.Plugin.Id, oldPluginRoute.TkeelVersion, resp.TkeelVersion)
		}
		addonsReq := &openapi_v1.AddonsIdentifyRequest{
			Plugin: &openapi_v1.BriefPluginInfo{
				Id:      resp.PluginId,
				Version: resp.Version,
			},
			ImplementedAddons: v.Addons,
		}
		addonsResp, err := s.openapiClient.AddonsIdentify(ctx, v.Plugin.Id, addonsReq)
		if err != nil {
			rbStack.Run()
			return nil, fmt.Errorf("error addons identify(%s/%s): %w", v.Plugin.Id, addonsReq, err)
		}
		if addonsResp.Res == nil {
			rbStack.Run()
			return nil, fmt.Errorf("error addons identify(%s/%s): Res is nil", v.Plugin.Id, addonsReq)
		}
		if addonsResp.Res.Ret != openapi_v1.Retcode_OK {
			rbStack.Run()
			return nil, fmt.Errorf("error addons identify(%s/%s): %s", v.Plugin.Id, addonsReq, addonsResp.Res.Msg)
		}
		pluginRouteBackup := oldPluginRoute.Clone()
		util.UpdatePluginRoute(resp.PluginId, v.Addons, oldPluginRoute)
		rbStack = append(rbStack, func() error {
			log.Debugf("register implemented plugin route roll back run.")
			err = s.pluginRouteOp.Update(ctx, pluginRouteBackup)
			if err != nil {
				return fmt.Errorf("error update plugin route backup(%s): %w", pluginRouteBackup, err)
			}
			return nil
		})
		err = s.pluginRouteOp.Update(ctx, oldPluginRoute)
		if err != nil {
			rbStack.Run()
			return nil, fmt.Errorf("error update old plugin route(%s): %w", oldPluginRoute, err)
		}
	}
	return rbStack, nil
}

func (s *PluginServiceV1) requestTenantEnable(ctx context.Context, pluginID string,
	tenantID string, extra []byte) (util.RollbackFunc, error) {
	resp, err := s.openapiClient.TenantEnable(ctx, pluginID, &openapi_v1.TenantEnableRequest{
		TenantId: tenantID,
		Extra:    extra,
	})
	if err != nil {
		log.Errorf("error tenant enalbe: %s", err)
		return nil, pb.PluginErrOpenapiEnabletenant()
	}
	if resp.Res == nil {
		log.Errorf("error tenant enalbe: res is nil")
		return nil, pb.PluginErrOpenapiEnabletenant()
	}
	if resp.Res.Ret != openapi_v1.Retcode_OK {
		log.Errorf("error tenant enalbe: %s", resp.Res.Msg)
		return nil, pb.PluginErrOpenapiEnabletenant()
	}
	return func() error {
		log.Debugf("roll back enable tenant: request(%s) disable(%s)", pluginID, tenantID)
		resp, err1 := s.openapiClient.TenantDisable(ctx, pluginID, &openapi_v1.TenantDisableRequest{
			TenantId: tenantID,
			Extra:    extra,
		})
		if err1 != nil {
			return fmt.Errorf("error request plugin(%s) tenant disable: %w", pluginID, err1)
		}
		if resp.Res == nil {
			return fmt.Errorf("error request plugin(%s) tenant disable: Res is nil", pluginID)
		}
		if resp.Res.Ret != openapi_v1.Retcode_OK {
			log.Errorf("error request plugin(%s) tenant enable: %s", pluginID, resp.Res.Msg)
		}
		return nil
	}, nil
}

func (s *PluginServiceV1) requestTenantDisable(ctx context.Context, pluginID string,
	tenantID string, extra []byte) (util.RollbackFunc, error) {
	resp, err := s.openapiClient.TenantDisable(ctx, pluginID, &openapi_v1.TenantDisableRequest{
		TenantId: tenantID,
		Extra:    extra,
	})
	if err != nil {
		log.Errorf("error tenant disable: %s", err)
		return nil, pb.PluginErrOpenapiDisableTenant()
	}
	if resp.Res == nil {
		log.Errorf("error tenant disable: res is nil")
		return nil, pb.PluginErrOpenapiDisableTenant()
	}
	if resp.Res.Ret != openapi_v1.Retcode_OK {
		log.Errorf("error tenant enalbe: %s", resp.Res.Msg)
		return nil, pb.PluginErrOpenapiDisableTenant()
	}
	return func() error {
		log.Debugf("roll back enable tenant: request(%s) enable(%s)", pluginID, tenantID)
		resp, err1 := s.openapiClient.TenantEnable(ctx, pluginID, &openapi_v1.TenantEnableRequest{
			TenantId: tenantID,
			Extra:    extra,
		})
		if err1 != nil {
			return fmt.Errorf("error request plugin(%s) tenant enable: %w", pluginID, err1)
		}
		if resp.Res == nil {
			return fmt.Errorf("error request plugin(%s) tenant enable: Res is nil", pluginID)
		}
		if resp.Res.Ret != openapi_v1.Retcode_OK {
			log.Errorf("error request plugin(%s) tenant enable: %s", pluginID, resp.Res.Msg)
		}
		return nil
	}, nil
}

func (s *PluginServiceV1) resetImplementedPluginRoute(ctx context.Context,
	unregisterPlugin *model.Plugin) (util.RollBackStack, error) {
	rbStack := util.NewRollbackStack()
	for _, v := range unregisterPlugin.ImplementedPlugin {
		oldRoute, err := s.pluginRouteOp.Get(ctx, v.Plugin.Id)
		if err != nil {
			rbStack.Run()
			return nil, fmt.Errorf("error plugin route(%s) get: %w", v.Plugin.Id, err)
		}
		tmpRoute := oldRoute.Clone()
		for _, a := range v.Addons {
			rbStack.Run()
			delete(oldRoute.RegisterAddons, a.AddonsPoint)
		}
		err = s.pluginRouteOp.Update(ctx, oldRoute)
		if err != nil {
			rbStack.Run()
			return nil, fmt.Errorf("error plugin route(%s) update: %w", oldRoute, err)
		}
		rbStack = append(rbStack, func() error {
			log.Debugf("delete plugin reset implemented plugin route roll back run.")
			if _, err := s.pluginRouteOp.Delete(ctx, tmpRoute.ID); err != nil {
				return fmt.Errorf("error delete tmpRoute(%s): %w", tmpRoute, err)
			}
			tmpRoute.Version = "1"
			if err := s.pluginRouteOp.Create(ctx, tmpRoute); err != nil {
				return fmt.Errorf("error create tmpRoute(%s): %w", tmpRoute, err)
			}
			return nil
		})
	}
	return rbStack, nil
}

func (s *PluginServiceV1) deletePlugin(ctx context.Context, pID string) (util.RollbackFunc, error) {
	dp, err := s.pluginOp.Delete(ctx, pID)
	if err != nil {
		log.Errorf("error delete plugin(%s): %s", pID, err)
		return nil, pb.PluginErrUninstallPlugin()
	}
	return func() error {
		log.Debugf("uninstall plugin delete plugin(%s) roll back run.", dp)
		dp.Version = "1"
		if err = s.pluginOp.Create(ctx, dp); err != nil {
			return fmt.Errorf("error roll back create plugin(%s): %w", dp, err)
		}
		return nil
	}, nil
}

func (s *PluginServiceV1) deletePluginRoute(ctx context.Context, pID string) (util.RollbackFunc, error) {
	dpr, err := s.pluginRouteOp.Delete(ctx, pID)
	if err != nil {
		if errors.Is(err, proute.ErrPluginRouteNotExsist) {
			return func() error {
				return nil
			}, nil
		}
		log.Errorf("error delete plugin(%s): %s", pID, err)
		return nil, pb.PluginErrUninstallPlugin()
	}
	return func() error {
		log.Debugf("uninstall plugin delete plugin(%s) route roll back run.", dpr)
		dpr.Version = "1"
		if err = s.pluginRouteOp.Create(ctx, dpr); err != nil {
			return fmt.Errorf("error roll back create plugin(%s) route: %w", dpr, err)
		}
		return nil
	}, nil
}

func getInstallerConfiguration(req *pb.InstallPluginRequest) (map[string]interface{}, error) {
	installerConfiguration := make(map[string]interface{})
	if req.Installer.Configuration != nil {
		switch req.Installer.Type {
		case pb.ConfigurationType_JSON:
			if err := json.Unmarshal(req.Installer.Configuration,
				&installerConfiguration); err != nil {
				return nil, fmt.Errorf("error unmarshal request installer info configuration: %w",
					err)
			}
		case pb.ConfigurationType_YAML:
			if err := yaml.Unmarshal(req.Installer.Configuration,
				&installerConfiguration); err != nil {
				return nil, fmt.Errorf("error unmarshal request installer info configuration: %w",
					err)
			}
		}
	}
	return installerConfiguration, nil
}

func convertConfiguration2Option(installerConfiguration map[string]interface{}) []*repository.Option {
	ret := make([]*repository.Option, 0, len(installerConfiguration))
	for k, v := range installerConfiguration {
		ret = append(ret, &repository.Option{
			Key:   k,
			Value: v,
		})
	}
	return ret
}

type enabledTenantList []*pb.EnabledTenant

func (a enabledTenantList) Len() int           { return len(a) }
func (a enabledTenantList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a enabledTenantList) Less(i, j int) bool { return a[i].EnableTimestamp < a[j].EnableTimestamp }

type pluginList []*pb.PluginBrief

func (a pluginList) Len() int      { return len(a) }
func (a pluginList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a pluginList) Less(i, j int) bool {
	if a[i].Id != a[j].Id {
		return a[i].Id < a[j].Id
	}
	iVer, err := strconv.ParseFloat(a[i].Version, 64)
	if err != nil {
		return true
	}
	jVer, err := strconv.ParseFloat(a[j].Version, 64)
	if err != nil {
		return false
	}
	return iVer < jVer
}
