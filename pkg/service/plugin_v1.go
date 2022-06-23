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
	"regexp"
	"sort"
	"time"

	"github.com/casbin/casbin/v2"
	g_version "github.com/hashicorp/go-version"
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
	"github.com/tkeel-io/tkeel/pkg/register"
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
	rbacOp         *casbin.SyncedEnforcer
}

func NewPluginServiceV1(rbacOp *casbin.SyncedEnforcer, db *gorm.DB, conf *config.TkeelConf, kvOp kv.Operator, pOp plugin.Operator,
	prOp proute.Operator, tpOp rbac.TenantPluginMgr, openapi openapi.Client,
) *PluginServiceV1 {
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
		rbacOp:         rbacOp,
	}
}

func (s *PluginServiceV1) InstallPlugin(ctx context.Context,
	req *pb.InstallPluginRequest,
) (*pb.InstallPluginResponse, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	if req.Installer == nil {
		log.Error("error install plugin request installer info is nil")
		return nil, pb.PluginErrInvalidArgument()
	}
	installerConfiguration, err := getInstallerConfiguration(req.Installer)
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
			return errors.Wrapf(err, "uninstall installer(%s)", installer.Brief())
		}
		return nil
	})
	// create new plugin.
	newP := model.NewPlugin(req.Id, &model.Installer{
		Repo:       installer.Brief().Repo,
		Name:       installer.Brief().Name,
		Version:    installer.Brief().Version,
		Icon:       installer.Brief().Icon,
		Desc:       installer.Brief().Desc,
		Maintainer: installer.Brief().Maintainers,
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
			return errors.Wrapf(err, "delete plugin(%s)", req.Id)
		}
		return nil
	})
	// tkeel tenant enalbe plugin.
	if _, err = s.tenantPluginOp.AddTenantPlugin(model.TKeelTenant, req.Id); err != nil {
		log.Errorf("error add tenant(%s) plugin(%s): %s", model.TKeelTenant, req.Id, err)
		return nil, pb.PluginErrUnknown()
	}
	rbStack = util.NewRollbackStack()
	register.Instance().Register(newP.ID, false, func() bool {
		actionCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
		defer cancel()
		return s.RegisterPluginAction(actionCtx, newP.ID, false)
	})
	log.Debugf("install plugin(%s) succ.", newP)
	return &pb.InstallPluginResponse{
		Plugin: util.ConvertModel2PluginObjectPb(newP, nil, model.TKeelTenant),
	}, nil
}

func (s *PluginServiceV1) UpgradePlugin(ctx context.Context,
	req *pb.UpgradePluginRequest,
) (*pb.UpgradePluginResponse, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	p, err := s.pluginOp.Get(ctx, req.GetId())
	if err != nil {
		log.Errorf("error get plugin(%s): %s", req.GetId(), err)
		if errors.Is(err, plugin.ErrPluginNotExsist) {
			return nil, pb.PluginErrPluginNotFound()
		}
		return nil, pb.PluginErrInternalStore()
	}
	if req.Installer == nil {
		log.Error("error upgrade plugin request installer info is nil")
		return nil, pb.PluginErrInvalidArgument()
	}
	installerConfiguration, err := getInstallerConfiguration(req.Installer)
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
	upgrader, err := repo.Get(req.Installer.Name, req.Installer.Version)
	if err != nil {
		log.Errorf("error get installer(%s): %s", req.Installer, err)
		return nil, pb.PluginErrInstallerNotFound()
	}
	upgrader.SetPluginID(req.Id)
	if err = upgrader.Upgrade(convertConfiguration2Option(installerConfiguration)...); err != nil {
		log.Errorf("error upgrade installer(%s) err: %s", upgrader.Brief(), err)
		if errors.Is(err, repository.ErrInvalidOptions) {
			return nil, pb.PluginErrInvalidArgument()
		}
		return nil, pb.PluginErrInstallInstaller()
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("installer roll back.")
		if err = hub.GetInstance().Uninstall(req.Id, upgrader.Brief()); err != nil {
			return errors.Wrapf(err, "uninstall installer(%s)", upgrader.Brief())
		}
		return nil
	})
	tmp := p.Clone()
	p.Upgrade(&model.Installer{
		Repo:       upgrader.Brief().Repo,
		Name:       upgrader.Brief().Name,
		Version:    upgrader.Brief().Version,
		Icon:       upgrader.Brief().Icon,
		Desc:       upgrader.Brief().Desc,
		Maintainer: upgrader.Brief().Maintainers,
	})
	rb, err := s.updatePlugin(ctx, tmp, p)
	if err != nil {
		log.Errorf("error update plugin(%s) err: %s", p, err)
		return nil, pb.PluginErrInternalStore()
	}
	rbStack = append(rbStack, rb)
	register.Instance().Register(p.ID, false, func() bool {
		actionCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
		defer cancel()
		return s.RegisterPluginAction(actionCtx, p.ID, false)
	})
	log.Debugf("upgrade plugin(%s) succ.", p)
	rbStack = util.NewRollbackStack()
	return &pb.UpgradePluginResponse{
		Plugin: util.ConvertModel2PluginObjectPb(p, nil, model.TKeelTenant),
	}, nil
}

func (s *PluginServiceV1) UninstallPlugin(ctx context.Context,
	req *pb.UninstallPluginRequest,
) (*pb.UninstallPluginResponse, error) {
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
		return nil, errors.Wrapf(err, "reset implemented plugin route(%s)", p.ID)
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
	rb, err = s.deleteTenantPluginEnable(ctx, req.Id)
	if err != nil {
		log.Errorf("error delete tenant(%s) plugin(%s): %s", model.TKeelTenant, req.Id, err)
		return nil, pb.PluginErrUnknown()
	}
	rbStack = append(rbStack, rb)
	// rbac remove plugin permission.
	rbList, err := s.deletePermission(ctx, p)
	if err != nil {
		log.Errorf("error delete permission: %s", req.GetId(), err)
		return nil, pb.PluginErrUninstallPlugin()
	}
	rbStack = append(rbStack, rbList...)
	// uninstall plugin.
	if err = hub.GetInstance().Uninstall(req.GetId(), &repository.InstallerBrief{
		Name:    p.Installer.Name,
		Repo:    p.Installer.Repo,
		Version: p.Installer.Version,
		State:   repository.StateInstalled,
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
	req *pb.GetPluginRequest,
) (*pb.GetPluginResponse, error) {
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
	req *pb.ListPluginRequest,
) (*pb.ListPluginResponse, error) {
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
	checkEnable := false
	if u.Tenant != model.TKeelTenant && !req.DisplayAllPlugin {
		checkEnable = true
	}
	for _, p := range ps {
		if checkEnable && p.DisableManualActivation {
			continue
		}
		pList = append(pList, util.ConvertModel2PluginBriefObjectPb(p, u.Tenant))
	}
	regular := getReglarStringKeyWords(req.KeyWords)
	exp, err := regexp.Compile(regular)
	if err != nil {
		log.Errorf("error compile regular(%s): %s", regular, err)
		return nil, pb.PluginErrInvalidArgument()
	}
	enableNum := 0
	ret := make(pluginList, 0, len(pList))
	for _, v := range pList {
		if exp.MatchString(v.Id) {
			ret = append(ret, v)
			if v.TenantEnable {
				enableNum++
			}
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
		EnableNum:  int32(enableNum),
	}, nil
}

func (s *PluginServiceV1) TenantEnable(ctx context.Context,
	req *pb.TenantEnableRequest,
) (*emptypb.Empty, error) {
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.PluginErrInvalidArgument()
	}
	_, terr, err := s.tenantEnablePlugin(ctx, false, u.Tenant, u.User, req.Id, req.Extra.Extra)
	if err != nil {
		log.Errorf("error tenant(%s) enable plugin(%s): %s", u.Tenant, u.User, err)
		return nil, terr
	}
	log.Debugf("tenant(%s) enable plugin(%s) succ.", u.Tenant, req.Id)
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) TMTenantEnable(ctx context.Context,
	req *pb.TMTenantEnableRequest,
) (*emptypb.Empty, error) {
	_, terr, err := s.tenantEnablePlugin(ctx, false, req.TenantId, model.TKeelUser, req.PluginId, req.Extra)
	if err != nil {
		log.Errorf("error tenant(%s) enable plugin(%s): %s", req.TenantId, model.TKeelUser, err)
		return nil, terr
	}
	log.Debugf("tenant(%s) enable plugin(%s) succ.", req.TenantId, req.PluginId)
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) TenantDisable(ctx context.Context,
	req *pb.TenantDisableRequest,
) (*emptypb.Empty, error) {
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

	tmpP := p.Clone()
	if p.TenantDisable(u.Tenant) {
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
		if _, err = s.updatePlugin(ctx, tmpP, p); err != nil {
			log.Errorf("error update tenant(%s) disable plugin(%s): %s", u.Tenant, p, err)
			return nil, pb.PluginErrInternalStore()
		}
	}
	rbStack = util.NewRollbackStack()
	log.Debugf("tenant(%s) disable plugin(%s) succ.", u.Tenant, req.Id)
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) TMTenantDisable(ctx context.Context,
	req *pb.TMTenantDisableRequest,
) (*emptypb.Empty, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	p, err := s.pluginOp.Get(ctx, req.PluginId)
	if err != nil {
		log.Errorf("error get plugin route: %s", err)
		return nil, pb.PluginErrInternalStore()
	}

	tmpP := p.Clone()
	if p.TenantDisable(req.TenantId) {
		// openapi tenant/disable.
		rb, err := s.requestTenantDisable(ctx, req.PluginId, req.TenantId, req.Extra)
		if err != nil {
			log.Errorf("error request(%s) tenant(%s/%s) disable: %s",
				req.PluginId, req.TenantId, string(req.Extra), err)
			return nil, pb.PluginErrOpenapiDisableTenant()
		}
		rbStack = append(rbStack, rb)
		// delete tenant plugin rbac.
		if _, err = s.tenantPluginOp.DeleteTenantPlugin(req.TenantId, req.PluginId); err != nil {
			log.Errorf("error delete tenant(%s) plugin(%s) rbac: %s", req.TenantId, p, err)
			return nil, pb.PluginErrUnknown()
		}
		// plugin update.
		if _, err = s.updatePlugin(ctx, tmpP, p); err != nil {
			log.Errorf("error update tenant(%s) disable plugin(%s): %s", req.TenantId, p, err)
			return nil, pb.PluginErrInternalStore()
		}
	}
	rbStack = util.NewRollbackStack()
	log.Debugf("tenant(%s) disable plugin(%s) succ.", req.TenantId, req.PluginId)
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) ListEnabledTenants(ctx context.Context,
	req *pb.ListEnabledTenantsRequest,
) (*pb.ListEnabledTenantsResponse, error) {
	p, err := s.pluginOp.Get(ctx, req.Id)
	if err != nil {
		log.Errorf("error get plugin: %s", err)
		if errors.Is(err, plugin.ErrPluginNotExsist) {
			return nil, pb.PluginErrInvalidArgument()
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

func (s *PluginServiceV1) TMUpdatePluginIdentify(ctx context.Context,
	req *pb.TMUpdatePluginIdentifyRequest,
) (*emptypb.Empty, error) {
	if req.Id != "" {
		if _, err := s.updatePluginIdentify(ctx, req.Id); err != nil {
			log.Errorf("error update plugin(%s) identify: %s", req.Id, err)
			return nil, pb.PluginErrInternalStore()
		}
	} else {
		ps, err := s.pluginOp.List(ctx)
		if err != nil {
			log.Errorf("error plugin op list: %s", err)
			return nil, pb.PluginErrInternalStore()
		}
		for _, v := range ps {
			_, err = s.updatePluginIdentify(ctx, v.ID)
			if err != nil {
				log.Errorf("error update identify(%s): %s", v.ID, err)
			}
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) TMRegisterPlugin(ctx context.Context,
	req *pb.TMRegisterPluginRequest,
) (*emptypb.Empty, error) {
	if req.Id == "" {
		log.Error("request plugin id is nil")
		return nil, pb.PluginErrInvalidArgument()
	}
	p, err := s.pluginOp.Get(ctx, req.Id)
	if err != nil {
		log.Errorf("error get plugin(%s): %s", req.Id, err)
		return nil, pb.PluginErrInternalStore()
	}
	if p.Status != openapi_v1.PluginStatus_ERR_REGISTER {
		log.Errorf("error plugin(%s) status not %s", req.Id, openapi_v1.PluginStatus_ERR_REGISTER)
		return nil, pb.PluginErrInvalidArgument()
	}
	if _, err = s.pluginRouteOp.Delete(ctx, req.Id); err != nil {
		log.Errorf("error delete plugin(%s) route: %s", req.Id, err)
		return nil, pb.PluginErrInternalStore()
	}
	if err = s.registerPluginProcess(ctx, req.Id, false); err != nil {
		log.Errorf("error plugin(%s) register: %s", req.Id, err)
		return nil, pb.PluginErrInternalStore()
	}
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) updatePluginIdentify(ctx context.Context, pID string) (util.RollbackFunc, error) {
	p, err := s.pluginOp.Get(ctx, pID)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin operator get %s", pID)
	}
	resp, err := s.queryIdentify(ctx, pID)
	if err != nil {
		return nil, errors.Wrapf(err, "query identify: %s", pID)
	}
	if err = s.checkIdentify(ctx, resp); err != nil {
		return nil, errors.Wrapf(err, "check identify: %s", resp)
	}
	oldP := p.Clone()
	p.Register(resp, helm.SecretContext)
	rb, err := s.updatePlugin(ctx, oldP, p)
	if err != nil {
		return nil, errors.Wrapf(err, "update plugin: %s", pID)
	}
	log.Debugf("update plugin(%s) identify", pID)
	return rb, nil
}

func (s *PluginServiceV1) RegisterPluginAction(ctx context.Context, pID string, isUpgrade bool) bool {
	resp, err := s.queryStatus(ctx, pID)
	if err != nil {
		log.Warnf("register query plugin(%s) status: %s", pID, err)
		return false
	} else if resp.Status == openapi_v1.PluginStatus_RUNNING {
		log.Debugf("register plugin(%s)", pID)
		// get register plugin identify.
		if err = s.registerPluginProcess(ctx, pID, isUpgrade); err != nil {
			log.Errorf("error register(%s): %s", pID, err)
			return false
		}
		log.Debugf("register plugin(%s) ok", pID)
		return true
	}
	return false
}

func (s *PluginServiceV1) registerPluginProcess(ctx context.Context, pID string, isUpgrade bool) error {
	resp, err := s.queryIdentify(ctx, pID)
	if err != nil {
		return errors.Wrap(err, "register error query identify")
	}
	// check register plugin identify.
	if err = s.checkIdentify(ctx, resp); err != nil {
		return errors.Wrap(err, "register error check identify")
	}
	if err = s.verifyPluginIdentity(ctx, resp, isUpgrade); err != nil {
		return errors.Wrap(err, "register error register plugin")
	}
	return nil
}

func (s *PluginServiceV1) updatePluginStatus(ctx context.Context, pID string, status openapi_v1.PluginStatus) error {
	// get plugin.
	p, err := s.pluginOp.Get(ctx, pID)
	if err != nil {
		return errors.Wrapf(err, "get plugin")
	}
	// update plugin status.
	p.Status = status
	if err = s.pluginOp.Update(ctx, p); err != nil {
		return errors.Wrapf(err, "update plugin")
	}
	return nil
}

func (s *PluginServiceV1) queryIdentify(ctx context.Context,
	pID string,
) (*openapi_v1.IdentifyResponse, error) {
	if pID == "" {
		return nil, errors.New("error empty plugin id")
	}
	identifyResp, err := s.openapiClient.Identify(ctx, pID)
	if err != nil {
		return nil, errors.Wrapf(err, "identify(%s)", pID)
	}
	if identifyResp.Res == nil {
		return nil, errors.Errorf("identify(%s): Res is nil", pID)
	}
	if identifyResp.Res.Ret != openapi_v1.Retcode_OK {
		return nil, errors.Errorf("identify(%s): %s", pID, identifyResp.Res.Ret)
	}
	if identifyResp.PluginId != pID {
		return nil, errors.Errorf("plugin id not match: %s -- %s", pID, identifyResp.PluginId)
	}
	return identifyResp, nil
}

func (s *PluginServiceV1) queryStatus(ctx context.Context, pID string) (*openapi_v1.StatusResponse, error) {
	if pID == "" {
		return nil, errors.New("error empty plugin id")
	}
	statusResp, err := s.openapiClient.Status(ctx, pID)
	if err != nil {
		return nil, errors.Wrapf(err, "status(%s)", pID)
	}
	if statusResp.Res == nil {
		return nil, errors.Errorf("status(%s): Res is nil", pID)
	}
	if statusResp.Res.Ret != openapi_v1.Retcode_OK {
		return nil, errors.Errorf("status(%s): %s", pID, statusResp.Res.Ret)
	}
	return statusResp, nil
}

func (s *PluginServiceV1) checkIdentify(ctx context.Context,
	resp *openapi_v1.IdentifyResponse,
) error {
	ok, err := util.CheckRegisterPluginTkeelVersion(resp.TkeelVersion, version.Version)
	if err != nil {
		return errors.Wrapf(err, "check register plugin(%s) depend tkeel version(%s)",
			resp.PluginId, resp.TkeelVersion)
	}
	if !ok {
		return errors.Errorf("plugin(%s) depend tkeel version(%s) not invalid",
			resp.PluginId, resp.TkeelVersion)
	}
	for _, v := range resp.Dependence {
		if pluginIsTkeelComponent(v.Id) {
			continue
		}
		if _, err := s.pluginOp.Get(ctx, v.Id); err != nil {
			return errors.Wrapf(err, "get dependence plugin(%s)", v.Id)
		}
	}
	return nil
}

func (s *PluginServiceV1) verifyPluginIdentity(ctx context.Context, resp *openapi_v1.IdentifyResponse, isUpgrade bool) error {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	if isUpgrade {
		// delete plugin route.
		rb, err := s.deletePluginRoute(ctx, resp.PluginId)
		if err != nil {
			return errors.Wrapf(err, "delete plugin route %s", resp.PluginId)
		}
		rbStack = append(rbStack, rb)
		// remove plugin permissions.
		rbs, err := util.DeletePluginPermissionOnSet(ctx, s.kvOp, resp.PluginId)
		if err != nil {
			return errors.Wrapf(err, "AddPluginPermissionOnSet %s", resp.PluginId)
		}
		rbStack = append(rbStack, rbs...)
	}
	// create plugin route.
	newPluginRoute := model.NewPluginRoute(resp)
	err := s.pluginRouteOp.Create(ctx, newPluginRoute)
	if err != nil {
		if errors.Is(err, proute.ErrPluginRouteExsist) {
			return ErrPluginRegistered
		}
		return errors.Wrapf(err, "create new plugin route(%s)", newPluginRoute)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("save register plugin route roll back run.")
		_, err = s.pluginRouteOp.Delete(ctx, newPluginRoute.ID)
		if err != nil {
			return errors.Wrapf(err, "delete new plugin(%s) route", newPluginRoute.ID)
		}
		return nil
	})
	// register implemented plugin route.
	rbs, err := s.checkImplementedPluginRoute(ctx, resp)
	if err != nil {
		return errors.Wrapf(err, "register implemented plugin(%s) route", resp.PluginId)
	}
	rbStack = append(rbStack, rbs...)
	// add plugin permissions.
	rbs, err = util.AddPluginPermissionOnSet(ctx, s.kvOp, resp.PluginId, resp.Permissions)
	if err != nil {
		return errors.Wrapf(err, "AddPluginPermissionOnSet %s", resp.PluginId)
	}
	rbStack = append(rbStack, rbs...)
	// update register plugin and update plugin route.
	p, err := s.pluginOp.Get(ctx, resp.PluginId)
	if err != nil {
		return errors.Wrapf(err, "get plugin(%s)", resp.PluginId)
	}
	p.Register(resp, helm.SecretContext)
	p.Status = openapi_v1.PluginStatus_RUNNING
	if err := s.pluginOp.Update(ctx, p); err != nil {
		return errors.Wrapf(err, "update plugin(%s)", p)
	}
	rbStack = util.NewRollbackStack()
	return nil
}

func (s *PluginServiceV1) checkImplementedPluginRoute(ctx context.Context,
	resp *openapi_v1.IdentifyResponse,
) (util.RollBackStack, error) {
	rbStack := util.NewRollbackStack()
	for _, v := range resp.ImplementedPlugin {
		oldPluginRoute, err := s.pluginRouteOp.Get(ctx, v.Plugin.Id)
		if err != nil {
			rbStack.Run()
			return nil, errors.Errorf("implemented plugin(%s) not registered", v.Plugin.Id)
		}
		ok, err := util.CheckRegisterPluginTkeelVersion(oldPluginRoute.TkeelVersion, resp.TkeelVersion)
		if err != nil {
			rbStack.Run()
			return nil, errors.Wrapf(err, "check implemented plugin(%s) depened tkeel version", v.Plugin.Id)
		}
		if !ok {
			rbStack.Run()
			return nil, errors.Errorf(`implemented plugin(%s) depened tkeel version(%s) 
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
			return nil, errors.Wrapf(err, "addons identify(%s/%s)", v.Plugin.Id, addonsReq)
		}
		if addonsResp.Res == nil {
			rbStack.Run()
			return nil, errors.Errorf("addons identify(%s/%s): Res is nil", v.Plugin.Id, addonsReq)
		}
		if addonsResp.Res.Ret != openapi_v1.Retcode_OK {
			rbStack.Run()
			return nil, errors.Errorf("addons identify(%s/%s): %s", v.Plugin.Id, addonsReq, addonsResp.Res.Msg)
		}
		pluginRouteBackup := oldPluginRoute.Clone()
		util.UpdatePluginRoute(resp.PluginId, v.Addons, oldPluginRoute)
		rbStack = append(rbStack, func() error {
			log.Debugf("register implemented plugin route roll back run.")
			err = s.pluginRouteOp.Update(ctx, pluginRouteBackup)
			if err != nil {
				return errors.Wrapf(err, "update plugin route backup(%s)", pluginRouteBackup)
			}
			return nil
		})
		err = s.pluginRouteOp.Update(ctx, oldPluginRoute)
		if err != nil {
			rbStack.Run()
			return nil, errors.Wrapf(err, "update old plugin route(%s)", oldPluginRoute)
		}
	}
	return rbStack, nil
}

func (s *PluginServiceV1) tenantEnablePlugin(ctx context.Context, isDependence bool,
	tenantID, userID, pluginID string, extra []byte,
) (util.RollBackStack, error, error) {
	rbStack := util.NewRollbackStack()
	p, err := s.pluginOp.Get(ctx, pluginID)
	if err != nil {
		return nil, pb.PluginErrInternalStore(), errors.Wrapf(err, "get plugin route")
	}
	if p.CheckTenantEnable(tenantID) {
		if !isDependence {
			return nil, pb.PluginErrDuplicateEnableTenant(), errors.Errorf("error tenant(%s) has been enabled", tenantID)
		}
		return rbStack, nil, nil
	}
	for _, v := range p.PluginDependences {
		if pluginIsTkeelComponent(v.Id) {
			continue
		}
		rbs, terr, err1 := s.tenantEnablePlugin(ctx, true, tenantID, userID, v.Id, extra)
		if err1 != nil {
			return nil, terr, errors.Wrapf(err1, "enant(%s) enable plugin(%s) enable dependence(%s)", tenantID, pluginID, v.Id)
		}
		rbStack = append(rbStack, rbs...)
	}
	// openapi tenant/enable.
	rb, err := s.requestTenantEnable(ctx, pluginID, tenantID, extra)
	if err != nil {
		rbStack.Run()
		return nil, pb.PluginErrOpenapiEnabletenant(), errors.Wrapf(err, "request(%s) tenant(%s/%s) enable",
			pluginID, tenantID, string(extra))
	}
	rbStack = append(rbStack, rb)
	// add tenant plugin rbac.
	if _, err = s.tenantPluginOp.AddTenantPlugin(tenantID, pluginID); err != nil {
		rbStack.Run()
		return nil, pb.PluginErrUnknown(), errors.Wrapf(err, "add tenant(%s) plugin(%s) rbac", tenantID, p)
	}
	// update plugin.
	tmpP := p.Clone()
	p.TenantEnable(&model.EnableTenant{
		TenantID:        tenantID,
		OperatorID:      userID,
		EnableTimestamp: time.Now().Unix(),
	})
	rb, err = s.updatePlugin(ctx, tmpP, p)
	if err != nil {
		rbStack.Run()
		return nil, pb.PluginErrInternalStore(), errors.Wrapf(err, "tenant(%s) enable(%s) update plugin", tenantID, p)
	}
	rbStack = append(rbStack, rb)
	return rbStack, nil, nil
}

func (s *PluginServiceV1) requestTenantEnable(ctx context.Context, pluginID string,
	tenantID string, extra []byte,
) (util.RollbackFunc, error) {
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
			return errors.Wrapf(err1, "request plugin(%s) tenant disable", pluginID)
		}
		if resp.Res == nil {
			return errors.Errorf("request plugin(%s) tenant disable: Res is nil", pluginID)
		}
		if resp.Res.Ret != openapi_v1.Retcode_OK {
			log.Errorf("error request plugin(%s) tenant enable: %s", pluginID, resp.Res.Msg)
		}
		return nil
	}, nil
}

func (s *PluginServiceV1) requestTenantDisable(ctx context.Context, pluginID string,
	tenantID string, extra []byte,
) (util.RollbackFunc, error) {
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
			return errors.Wrapf(err1, "request plugin(%s) tenant enable", pluginID)
		}
		if resp.Res == nil {
			return errors.Errorf("request plugin(%s) tenant enable: Res is nil", pluginID)
		}
		if resp.Res.Ret != openapi_v1.Retcode_OK {
			log.Errorf("error request plugin(%s) tenant enable: %s", pluginID, resp.Res.Msg)
		}
		return nil
	}, nil
}

func (s *PluginServiceV1) resetImplementedPluginRoute(ctx context.Context,
	unregisterPlugin *model.Plugin,
) (util.RollBackStack, error) {
	rbStack := util.NewRollbackStack()
	for _, v := range unregisterPlugin.ImplementedPlugin {
		oldRoute, err := s.pluginRouteOp.Get(ctx, v.Plugin.Id)
		if err != nil {
			rbStack.Run()
			return nil, errors.Wrapf(err, "plugin route(%s) get", v.Plugin.Id)
		}
		tmpRoute := oldRoute.Clone()
		for _, a := range v.Addons {
			rbStack.Run()
			delete(oldRoute.RegisterAddons, a.AddonsPoint)
		}
		err = s.pluginRouteOp.Update(ctx, oldRoute)
		if err != nil {
			rbStack.Run()
			return nil, errors.Wrapf(err, "plugin route(%s) update", oldRoute)
		}
		rbStack = append(rbStack, func() error {
			log.Debugf("delete plugin reset implemented plugin route roll back run.")
			if _, err := s.pluginRouteOp.Delete(ctx, tmpRoute.ID); err != nil {
				return errors.Wrapf(err, "delete tmpRoute(%s)", tmpRoute)
			}
			tmpRoute.Version = "1"
			if err := s.pluginRouteOp.Create(ctx, tmpRoute); err != nil {
				return errors.Wrapf(err, "create tmpRoute(%s)", tmpRoute)
			}
			return nil
		})
	}
	return rbStack, nil
}

func (s *PluginServiceV1) updatePlugin(ctx context.Context, oldP, newP *model.Plugin) (util.RollbackFunc, error) {
	if err := s.pluginOp.Update(ctx, newP); err != nil {
		return nil, errors.Wrapf(err, "update plugin(%s)", newP)
	}
	return func() error {
		log.Debugf("roll back update plugin: %s --> %s", newP, oldP)
		if _, err := s.pluginOp.Delete(ctx, oldP.ID); err != nil {
			log.Errorf("error roll back update plugin(%s) delete: %s", oldP, err)
			return errors.Wrap(err, "plugin delete")
		}
		oldP.Version = "1"
		if err := s.pluginOp.Create(ctx, oldP); err != nil {
			log.Errorf("error roll back update plugin(%s) create: %s", oldP, err)
			return errors.Wrap(err, "plugin create")
		}
		return nil
	}, nil
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
			return errors.Wrapf(err, "roll back create plugin(%s)", dp)
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
			return errors.Wrapf(err, "roll back create plugin(%s) route", dpr)
		}
		return nil
	}, nil
}

func (s *PluginServiceV1) deleteTenantPluginEnable(ctx context.Context, pID string) (util.RollbackFunc, error) {
	if _, err := s.tenantPluginOp.DeleteTenantPlugin(model.TKeelTenant, pID); err != nil {
		log.Errorf("error delete tenant(%s) plugin(%s): %s", model.TKeelTenant, pID, err)
		return nil, pb.PluginErrUnknown()
	}
	return func() error {
		log.Debugf("uninstall plugin  tenant(%s) plugin(%s) route roll back run.", model.TKeelTenant, pID)
		if _, err := s.tenantPluginOp.AddTenantPlugin(model.TKeelTenant, pID); err != nil {
			return errors.Wrapf(err, "add tenant(%s) plugin(%s)", model.TKeelTenant, pID)
		}
		return nil
	}, nil
}

func (s *PluginServiceV1) deletePermission(ctx context.Context, p *model.Plugin) (util.RollBackStack, error) {
	rbStack := util.NewRollbackStack()
	pList := model.GetPermissionSet().GetAllPermissionByPluginID(p.ID)
	removePolicies := make([][]string, 0)
	for _, v := range pList {
		removePolicies = append(removePolicies,
			s.rbacOp.GetFilteredPolicy(2, v.Path, model.AllowedPermissionAction)...)
	}
	if _, err := s.rbacOp.RemovePolicies(removePolicies); err != nil {
		return nil, errors.Wrapf(err, "RemovePolicies %v", removePolicies)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("deletePermission %s RemovePolicies roll back run", p.ID)
		if _, err := s.rbacOp.AddPolicies(removePolicies); err != nil {
			return errors.Wrapf(err, "AddPolicies %v", removePolicies)
		}
		return nil
	})
	model.GetPermissionSet().Delete(p.ID)
	rbStack = append(rbStack, func() error {
		log.Debugf("deletePermission %s Delete permission set roll back run", p.ID)
		if _, err := util.AddPluginPermissionOnSet(ctx, s.kvOp, p.ID, p.Permissions); err != nil {
			return errors.Wrapf(err, "AddPluginPermissionOnSet %s", p.ID)
		}
		return nil
	})
	return rbStack, nil
}

func getInstallerConfiguration(reqInstaller *pb.Installer) (map[string]interface{}, error) {
	installerConfiguration := make(map[string]interface{})
	if reqInstaller.Configuration != nil {
		switch reqInstaller.Type {
		case pb.ConfigurationType_JSON:
			if err := json.Unmarshal(reqInstaller.Configuration,
				&installerConfiguration); err != nil {
				return nil, errors.Wrap(err, "unmarshal request installer info configuration")
			}
		case pb.ConfigurationType_YAML:
			if err := yaml.Unmarshal(reqInstaller.Configuration,
				&installerConfiguration); err != nil {
				return nil, errors.Wrap(err, "unmarshal request installer info configuration")
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
	iVer, err := g_version.NewVersion(a[i].Version)
	if err != nil {
		return true
	}
	jVer, err := g_version.NewVersion(a[j].Version)
	if err != nil {
		return false
	}
	if !iVer.Equal(jVer) {
		return iVer.LessThan(jVer)
	}
	return a[i].RegisterTimestamp < a[j].RegisterTimestamp
}
