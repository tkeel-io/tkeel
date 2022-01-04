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

	"github.com/pkg/errors"
	t_errors "github.com/tkeel-io/kit/errors"
	"github.com/tkeel-io/kit/log"
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
	"github.com/tkeel-io/tkeel/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/yaml.v3"
)

var (
	ErrGetOpenapiIdentify = errors.New("error get openapi identify")
	ErrPluginRegistered   = errors.New("plugin is registered")
)

type PluginServiceV1 struct {
	pb.UnimplementedPluginServer

	tkeelConf     *config.TkeelConf
	kvOp          kv.Operator
	pluginOp      plugin.Operator
	pluginRouteOp proute.Operator
	openapiClient openapi.Client
}

func NewPluginServiceV1(conf *config.TkeelConf, kvOp kv.Operator, pOp plugin.Operator,
	prOp proute.Operator, openapi openapi.Client) *PluginServiceV1 {
	return &PluginServiceV1{
		tkeelConf:     conf,
		kvOp:          kvOp,
		pluginOp:      pOp,
		pluginRouteOp: prOp,
		openapiClient: openapi,
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
	installerConfiguration := make(map[string]interface{})
	if req.Installer.Configuration != nil {
		switch req.Installer.Type {
		case pb.ConfigurationType_JSON:
			if err := json.Unmarshal(req.Installer.Configuration,
				&installerConfiguration); err != nil {
				log.Errorf("error unmarshal request installer info configuration: %s",
					err)
				return nil, pb.PluginErrInvalidArgument()
			}
		case pb.ConfigurationType_YAML:
			if err := yaml.Unmarshal(req.Installer.Configuration,
				&installerConfiguration); err != nil {
				log.Errorf("error unmarshal request installer info configuration: %s",
					err)
				return nil, pb.PluginErrInvalidArgument()
			}
		}
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
	if err = installer.Install(func() []*repository.Option {
		ret := make([]*repository.Option, 0, len(installerConfiguration))
		for k, v := range installerConfiguration {
			ret = append(ret, &repository.Option{
				Key:   k,
				Value: v,
			})
		}
		return ret
	}()...); err != nil {
		log.Errorf("error install installer(%s) err: %s", installer.Brief(), err)
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
	newP := model.NewPlugin(req.Id, &model.Installer{
		Repo:    installer.Brief().Repo,
		Name:    installer.Brief().Name,
		Version: installer.Brief().Version,
	})
	if err = s.pluginOp.Create(ctx, newP); err != nil {
		log.Errorf("error create plugin(%s): %s", newP, err)
		if errors.Is(err, plugin.ErrPluginExsist) {
			return nil, pb.PluginErrPluginAlreadyExists()
		}
		return nil, pb.PluginErrInternalStore()
	}
	rbStack = util.NewRollbackStack()
	log.Debugf("install plugin(%s) succ.", newP)
	return &pb.InstallPluginResponse{
		Plugin: util.ConvertModel2PluginObjectPb(newP, nil),
	}, nil
}

func (s *PluginServiceV1) UninstallPlugin(ctx context.Context,
	req *pb.UninstallPluginRequest) (*pb.UninstallPluginResponse, error) {
	mp, err := s.pluginOp.Get(ctx, req.GetId())
	if err != nil {
		log.Errorf("error plugin operator get: %s", err)
		if errors.Is(err, plugin.ErrPluginNotExsist) {
			return nil, pb.PluginErrPluginNotFound()
		}
		return nil, pb.PluginErrInternalStore()
	}
	if mp.Installer == nil {
		log.Errorf("error plugin(%s) installer is nil", mp)
		return nil, pb.PluginErrInternalStore()
	}
	if err = hub.GetInstance().Uninstall(req.GetId(), &repository.InstallerBrief{
		Name:      mp.Installer.Name,
		Repo:      mp.Installer.Repo,
		Version:   mp.Installer.Version,
		Installed: true,
	}); err != nil {
		log.Errorf("error uninstall plugin(%s): %s", mp, err)
		return nil, pb.PluginErrUninstallPlugin()
	}
	// delete plugin.
	dp, err := s.pluginOp.Delete(ctx, req.GetId())
	if err != nil {
		log.Errorf("error delete plugin(%s): %s", req.GetId(), err)
		return nil, pb.PluginErrUninstallPlugin()
	}
	log.Debugf("uninstall plugin(%s) succ.", dp)
	return &pb.UninstallPluginResponse{
		Plugin: util.ConvertModel2PluginObjectPb(mp, nil),
	}, nil
}

func (s *PluginServiceV1) RegisterPlugin(ctx context.Context,
	req *pb.RegisterPluginRequest) (retResp *emptypb.Empty, err error) {
	// get plugin.
	mp, err := s.pluginOp.Get(ctx, req.GetId())
	if err != nil {
		log.Errorf("error get plugin: %s", err)
		if errors.Is(err, plugin.ErrPluginNotExsist) {
			return nil, pb.PluginErrPluginNotFound()
		}
		return nil, pb.PluginErrInternalStore()
	}
	// get register plugin identify.
	resp, err := s.queryIdentify(ctx, req.GetId())
	if err != nil {
		log.Errorf("error query identify: %s", err)
		return nil, pb.PluginErrInvalidArgument()
	}
	// check register plugin identify.
	if err = s.checkIdentify(ctx, resp); err != nil {
		log.Errorf("error check identify: %s", err)
		return nil, pb.PluginErrInvalidArgument()
	}
	// register plugin.
	if err = s.registerPlugin(ctx, mp, req.GetSecret(), resp); err != nil {
		log.Errorf("error register plugin: %s", err)
		if errors.Is(err, ErrPluginRegistered) {
			return nil, pb.PluginErrPluginAlreadyExists()
		}
		return nil, pb.PluginErrInternalQueryPluginOpenapi()
	}
	log.Debugf("register plugin(%s) ok", mp)
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) UnregisterPlugin(ctx context.Context,
	req *pb.UnregisterPluginRequest) (*pb.UnregisterPluginResponse, error) {
	pID := req.Id
	// check exists.
	UnregisterPluginRoute, err := s.pluginRouteOp.Get(ctx, pID)
	if err != nil {
		log.Errorf("error unregister plugin(%s) route get: %s", pID, err)
		if errors.Is(err, proute.ErrPluginRouteNotExsist) {
			return nil, pb.PluginErrPluginNotFound()
		}
		return nil, pb.PluginErrInternalStore()
	}
	// check whether the extension point is implemented.
	if len(UnregisterPluginRoute.RegisterAddons) != 0 {
		log.Errorf("error unregister plugin(%s): other plugins have implemented the extension points of this plugin.", pID)
		return nil, pb.PluginErrUnregisterPluginHasBeenDepended()
	}

	// delete plugin.
	unregisterPlugin, delPluginRoute, err := s.deletePluginRoute(ctx, pID)
	if err != nil {
		log.Errorf("error delete plugin(%s) route: %s", pID, err)
		return nil, pb.PluginErrInternalStore()
	}
	log.Debugf("unregister plugin(%s) succ.", unregisterPlugin)
	return &pb.UnregisterPluginResponse{
		Plugin: util.ConvertModel2PluginObjectPb(unregisterPlugin, delPluginRoute),
	}, nil
}

func (s *PluginServiceV1) GetPlugin(ctx context.Context,
	req *pb.GetPluginRequest) (*pb.GetPluginResponse, error) {
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
		if errors.Is(err, proute.ErrPluginRouteNotExsist) {
			return &pb.GetPluginResponse{
				Plugin: util.ConvertModel2PluginObjectPb(gPlugin, nil),
			}, nil
		}
		log.Errorf("error plugin(%s) route get: %s", req.Id, err)
		return nil, pb.PluginErrInternalStore()
	}

	return &pb.GetPluginResponse{
		Plugin: util.ConvertModel2PluginObjectPb(gPlugin, gPluginRoute),
	}, nil
}

func (s *PluginServiceV1) ListPlugin(ctx context.Context,
	req *emptypb.Empty) (*pb.ListPluginResponse, error) {
	ps, err := s.pluginOp.List(ctx)
	if err != nil {
		log.Errorf("error plugin list: %s", err)
		return nil, pb.PluginErrListPlugin()
	}
	retList := make([]*pb.PluginObject, 0, len(ps))
	for _, p := range ps {
		var pbPlugin *pb.PluginObject
		if p.State == openapi_v1.PluginStatus_UNREGISTER {
			pbPlugin = util.ConvertModel2PluginObjectPb(p, nil)
		} else {
			pr, err := s.pluginRouteOp.Get(ctx, p.ID)
			if err != nil {
				log.Errorf("error plugin list get plugin(%s) route: %s", p.ID, err)
				return nil, pb.PluginErrInternalStore()
			}
			pbPlugin = util.ConvertModel2PluginObjectPb(p, pr)
		}
		retList = append(retList, pbPlugin)
	}

	return &pb.ListPluginResponse{
		PluginList: retList,
	}, nil
}

func (s *PluginServiceV1) BindTenants(ctx context.Context,
	req *pb.BindTenantsRequest) (*emptypb.Empty, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	u, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.PluginErrInvalidArgument()
	}
	pr, err := s.pluginRouteOp.Get(ctx, req.Id)
	if err != nil {
		log.Errorf("error get plugin route: %s", err)
		return nil, pb.PluginErrInternalStore()
	}
	tmpPr := pr.Clone()
	for _, v := range pr.ActiveTenantes {
		if v == u.Tenant {
			log.Errorf("error plugin(%s) duplicat tenant tenant(%s)", req.Id, v)
			return nil, pb.PluginErrDuplicateActiveTenant()
		}
	}
	// openapi tenant/bind.
	rb, err := s.requestTenantBind(ctx, req.Id, u.Tenant, req.Extra)
	if err != nil {
		log.Errorf("error request(%s) tenant(%s/%s) bind: %s",
			req.Id, u.Tenant, string(req.Extra), err)
		if errors.As(err, t_errors.TError{}) {
			return nil, err
		}
		return nil, pb.PluginErrUnknown()
	}
	rbStack = append(rbStack, rb)
	// TODO: security store.
	// update plugin route.
	pr.ActiveTenantes = append(pr.ActiveTenantes, u.Tenant)
	if err = s.pluginRouteOp.Update(ctx, pr); err != nil {
		log.Errorf("error bind tenant(%s) update plugin(%s) route", u.Tenant, req.Id)
		return nil, pb.PluginErrInternalStore()
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("roll back bind tenant: %s --> %s", pr, tmpPr)
		if _, err = s.pluginRouteOp.Delete(ctx, tmpPr.ID); err != nil {
			log.Errorf("error roll back update plugin(%s) route delete: %s", tmpPr, err)
			return fmt.Errorf("error pr delete: %w", err)
		}
		if err = s.pluginRouteOp.Create(ctx, tmpPr); err != nil {
			log.Errorf("error roll back update plugin(%s) route create: %s", tmpPr, err)
			return fmt.Errorf("error pr create: %w", err)
		}
		return nil
	})
	// tenant bind model.
	tbKey := model.GetTenantBindKey(u.Tenant)
	vsb, ver, err := s.kvOp.Get(ctx, tbKey)
	if err != nil {
		log.Errorf("error get tenant(%s) bind: %s", u.Tenant, err)
		return nil, pb.PluginErrInternalStore()
	}
	tbBinds := model.ParseTenantBind(vsb)
	tbBinds = append(tbBinds, req.Id)
	newValue := model.EncodeTenantBind(tbBinds)
	if ver == "" {
		if err = s.kvOp.Create(ctx, tbKey, newValue); err != nil {
			log.Errorf("error create new(%s) tenant bind(%s): %s", tbKey, string(newValue), err)
			return nil, pb.PluginErrInternalStore()
		}
	} else {
		if err = s.kvOp.Update(ctx, tbKey, newValue, ver); err != nil {
			log.Errorf("error update new(%s) tenant bind(%s): %s", tbKey, string(newValue), err)
			return nil, pb.PluginErrInternalStore()
		}
	}
	rbStack = util.NewRollbackStack()
	log.Debugf("plugin(%s) bind(%s) succ.", req.Id, u.Tenant)
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) requestTenantBind(ctx context.Context, pluginID string,
	tenantID string, extra []byte) (util.RollbackFunc, error) {
	resp, err := s.openapiClient.TenantBind(ctx, pluginID, &openapi_v1.TenantBindRequst{
		TenantId: tenantID,
		Extra:    extra,
	})
	if err != nil {
		return nil, pb.PluginErrOpenapiBindtenant()
	}
	if resp.Res.Ret != openapi_v1.Retcode_OK {
		return nil, pb.PluginErrOpenapiBindtenant()
	}
	return func() error {
		log.Debugf("roll back bind tenant: request(%s) unbind(%s)", pluginID, tenantID)
		resp, err1 := s.openapiClient.TenantUnbind(ctx, pluginID, &openapi_v1.TenantUnbindRequst{
			TenantId: tenantID,
			Extra:    extra,
		})
		if err1 != nil {
			return fmt.Errorf("error request plugin(%s) tenant unbind: %w", pluginID, err1)
		}
		if resp.Res.Ret != openapi_v1.Retcode_OK {
			log.Errorf("error request plugin(%s) tenant bind: %s", pluginID, resp.Res.Msg)
		}
		return nil
	}, nil
}

func (s *PluginServiceV1) UnbindTenants(ctx context.Context,
	req *pb.UnbindTenantsRequest) (*emptypb.Empty, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	user, err := util.GetUser(ctx)
	if err != nil {
		log.Errorf("error get user: %s", err)
		return nil, pb.PluginErrInvalidArgument()
	}
	pr, err := s.pluginRouteOp.Get(ctx, req.Id)
	if err != nil {
		log.Errorf("error get plugin route: %s", err)
		return nil, pb.PluginErrInternalStore()
	}
	update := false
	for i, v := range pr.ActiveTenantes {
		if v == user.Tenant {
			pr.ActiveTenantes = append(pr.ActiveTenantes[:i], pr.ActiveTenantes[i+1:]...)
			update = true
			break
		}
	}
	if update {
		// openapi tenant/unbind.
		rb, err := s.requestTenantUnbind(ctx, req.Id, user.Tenant, req.Extra)
		if err != nil {
			log.Errorf("error request(%s) tenant(%s/%s) unbind: %s",
				req.Id, user.Tenant, string(req.Extra), err)
			if errors.As(err, t_errors.TError{}) {
				return nil, err
			}
			return nil, pb.PluginErrUnknown()
		}
		rbStack = append(rbStack, rb)
		// TODO: security store.
		// plugin route update.
		if err = s.pluginRouteOp.Update(ctx, pr); err != nil {
			log.Errorf("error unbind tenant(%s) update plugin(%s) route", user.Tenant, req.Id)
			return nil, pb.PluginErrInternalStore()
		}
		tmpPr := pr.Clone()
		rbStack = append(rbStack, func() error {
			log.Debugf("roll back unbind tenant: %s --> %s", pr, tmpPr)
			if _, err = s.pluginRouteOp.Delete(ctx, tmpPr.ID); err != nil {
				log.Errorf("error roll back update plugin(%s) route delete: %s", tmpPr, err)
				return fmt.Errorf("error pr delete: %w", err)
			}
			if err = s.pluginRouteOp.Create(ctx, tmpPr); err != nil {
				log.Errorf("error roll back update plugin(%s) route create: %s", tmpPr, err)
				return fmt.Errorf("error pr create: %w", err)
			}
			return nil
		})
		// tenant bind update.
		tbKey := model.GetTenantBindKey(user.Tenant)
		vsb, ver, err := s.kvOp.Get(ctx, tbKey)
		if err != nil {
			log.Errorf("error get tenant(%s) bind:cd  %s", user.Tenant, err)
			return nil, pb.PluginErrInternalStore()
		}
		tbBinds := model.ParseTenantBind(vsb)
		for i, v := range tbBinds {
			if v == req.Id {
				tbBinds = append(tbBinds[:i], tbBinds[i+1:]...)
				break
			}
		}
		newValue := model.EncodeTenantBind(tbBinds)
		if err = s.kvOp.Update(ctx, tbKey, newValue, ver); err != nil {
			log.Errorf("error update new(%s) tenant bind(%s): %s", tbKey, string(newValue), err)
			return nil, pb.PluginErrInternalStore()
		}
	}
	rbStack = util.NewRollbackStack()
	log.Debugf("plugin(%s) unbind(%s) succ.", req.Id, user.Tenant)
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) requestTenantUnbind(ctx context.Context, pluginID string,
	tenantID string, extra []byte) (util.RollbackFunc, error) {
	resp, err := s.openapiClient.TenantUnbind(ctx, pluginID, &openapi_v1.TenantUnbindRequst{
		TenantId: tenantID,
		Extra:    extra,
	})
	if err != nil {
		return nil, pb.PluginErrOpenapiBindtenant()
	}
	if resp.Res.Ret != openapi_v1.Retcode_OK {
		return nil, pb.PluginErrOpenapiBindtenant()
	}
	return func() error {
		log.Debugf("roll back bind tenant: request(%s) bind(%s)", pluginID, tenantID)
		resp, err1 := s.openapiClient.TenantBind(ctx, pluginID, &openapi_v1.TenantBindRequst{
			TenantId: tenantID,
			Extra:    extra,
		})
		if err1 != nil {
			return fmt.Errorf("error request plugin(%s) tenant bind: %w", pluginID, err1)
		}
		if resp.Res.Ret != openapi_v1.Retcode_OK {
			log.Errorf("error request plugin(%s) tenant bind: %s", pluginID, resp.Res.Msg)
		}
		return nil
	}, nil
}

func (s *PluginServiceV1) ListBindTenants(ctx context.Context,
	req *pb.ListBindTenantsRequest) (*pb.ListBindTenantsResponse, error) {
	pr, err := s.pluginRouteOp.Get(ctx, req.Id)
	if err != nil {
		log.Errorf("error get plugin route: %s", err)
		if errors.Is(err, proute.ErrPluginRouteNotExsist) {
			return nil, pb.PluginErrPluginNotFound()
		}
		return nil, pb.PluginErrInternalStore()
	}
	return &pb.ListBindTenantsResponse{
		Tenants: pr.ActiveTenantes,
	}, nil
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
	if identifyResp.Res.Ret != openapi_v1.Retcode_OK {
		return nil, fmt.Errorf("error identify(%s): %s", pID, identifyResp.Res.Ret)
	}
	if identifyResp.PluginId != pID {
		return nil, fmt.Errorf("error plugin id not match: %s -- %s", pID, identifyResp.PluginId)
	}
	return identifyResp, nil
}

func (s *PluginServiceV1) checkIdentify(ctx context.Context,
	resp *openapi_v1.IdentifyResponse) error {
	ok, err := util.CheckRegisterPluginTkeelVersion(resp.TkeelVersion, s.tkeelConf.Version)
	if err != nil {
		return fmt.Errorf("error check register plugin(%s) depend tkeel version(%s): %w",
			resp.PluginId, resp.TkeelVersion, err)
	}
	if !ok {
		return fmt.Errorf("error plugin(%s) depend tkeel version(%s) not invalid",
			resp.PluginId, resp.TkeelVersion)
	}
	return nil
}

func (s *PluginServiceV1) registerPlugin(ctx context.Context, registerPlugin *model.Plugin,
	secret string, resp *openapi_v1.IdentifyResponse) error {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	// save register plugin route.
	rb, tmpPluginRoute, err := s.saveRegisterPluginRouter(ctx, resp)
	if err != nil {
		if errors.Is(err, ErrPluginRegistered) {
			return ErrPluginRegistered
		}
		return fmt.Errorf("error save register plugin(%s) route: %w", resp.PluginId, err)
	}
	rbStack = append(rbStack, rb)
	// register implemented plugin route.
	rbs, err := s.registerImplementedPluginRoute(ctx, resp)
	if err != nil {
		return fmt.Errorf("error register implemented plugin(%s) route: %w", resp.PluginId, err)
	}
	rbStack = append(rbStack, rbs...)
	// update register plugin and update plugin route.
	err = s.updateRegisterPlugin(ctx, resp, secret, registerPlugin, tmpPluginRoute)
	if err != nil {
		return fmt.Errorf("error update register plugin(%s): %w", resp.PluginId, err)
	}
	rbStack = util.NewRollbackStack()
	return nil
}

func (s *PluginServiceV1) saveRegisterPluginRouter(ctx context.Context,
	resp *openapi_v1.IdentifyResponse) (util.RollbackFunc, *model.PluginRoute, error) {
	// create plugin route.
	newPluginRoute := model.NewPluginRoute(resp)
	err := s.pluginRouteOp.Create(ctx, newPluginRoute)
	if err != nil {
		if errors.Is(err, proute.ErrPluginRouteExsist) {
			return nil, nil, ErrPluginRegistered
		}
		return nil, nil, fmt.Errorf("error create new plugin route(%s): %w", newPluginRoute, err)
	}
	return func() error {
		log.Debugf("save register plugin route roll back run.")
		_, err := s.pluginRouteOp.Delete(ctx, newPluginRoute.ID)
		if err != nil {
			return fmt.Errorf("error delete new plugin(%s) route: %w", newPluginRoute.ID, err)
		}
		return nil
	}, newPluginRoute, nil
}

func (s *PluginServiceV1) registerImplementedPluginRoute(ctx context.Context,
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

func (s *PluginServiceV1) updateRegisterPlugin(ctx context.Context, resp *openapi_v1.IdentifyResponse,
	secret string, oldPlugin *model.Plugin, tmpPluginRoute *model.PluginRoute) error {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	var state openapi_v1.PluginStatus
	statusResp, err := s.openapiClient.Status(ctx, resp.PluginId)
	if err != nil {
		return fmt.Errorf("error status(%s): %w", resp.PluginId, err)
	}
	if statusResp.Res.Ret != openapi_v1.Retcode_OK {
		state = openapi_v1.PluginStatus_ERROR
	} else {
		state = statusResp.Status
	}
	tmpPlugin := oldPlugin.Clone()
	oldPlugin.Register(resp, secret)
	oldPlugin.State = state
	err = s.pluginOp.Update(ctx, oldPlugin)
	if err != nil {
		return fmt.Errorf("error update plugin(%s): %w", oldPlugin, err)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("update register plugin roll back run.")
		_, err = s.pluginOp.Delete(ctx, tmpPlugin.ID)
		if err != nil {
			return fmt.Errorf("error delete oldPlugin(%s): %w", tmpPlugin, err)
		}
		tmpPlugin.Version = "1"
		if err = s.pluginOp.Create(ctx, tmpPlugin); err != nil {
			return fmt.Errorf("error create oldPlugin(%s): %w", tmpPlugin, err)
		}
		return nil
	})
	tmpPluginRoute.Status = state
	err = s.pluginRouteOp.Update(ctx, tmpPluginRoute)
	if err != nil {
		return fmt.Errorf("error update tmp plugin route(%s): %w", tmpPluginRoute, err)
	}
	rbStack = util.NewRollbackStack()
	return nil
}

func (s *PluginServiceV1) deletePluginRoute(ctx context.Context, deleteID string) (*model.Plugin, *model.PluginRoute, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	// delete plugin route.
	deletePluginRoute, err := s.pluginRouteOp.Delete(ctx, deleteID)
	if err != nil {
		return nil, nil, fmt.Errorf("error unregister delete plugin(%s) route: %w", deleteID, err)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("unregister delete plugin route roll back run.")
		err = s.pluginRouteOp.Create(ctx, deletePluginRoute)
		if err != nil {
			return fmt.Errorf("error create delete plugin(%s) route: %w", deletePluginRoute, err)
		}
		return nil
	})
	// reset implemented plugin route.
	unregisterPlugin, err := s.pluginOp.Get(ctx, deleteID)
	if err != nil {
		return nil, nil, fmt.Errorf("error unregister plugin(%s): %w", deleteID, err)
	}
	subRbStack, err := s.resetImplementedPluginRoute(ctx, unregisterPlugin)
	if err != nil {
		return nil, nil, fmt.Errorf("error reset implemented plugin route(%s): %w", deleteID, err)
	}
	rbStack = append(rbStack, subRbStack...)
	// update plugin.
	unregisterPlugin.State = openapi_v1.PluginStatus_UNREGISTER
	if err = s.pluginOp.Update(ctx, unregisterPlugin); err != nil {
		return nil, nil, fmt.Errorf("error update unregister plugin(%s): %w", unregisterPlugin, err)
	}
	oldPlugin := unregisterPlugin.Clone()
	rbStack = append(rbStack, func() error {
		log.Debugf("unregister plugin roll back run.")
		_, err = s.pluginOp.Delete(ctx, oldPlugin.ID)
		if err != nil {
			return fmt.Errorf("error delete oldPlugin(%s): %w", oldPlugin, err)
		}
		oldPlugin.Version = "1"
		err = s.pluginOp.Create(ctx, oldPlugin)
		if err != nil {
			return fmt.Errorf("error create unregister plugin(%s): %w", oldPlugin, err)
		}
		return nil
	})

	rbStack = util.NewRollbackStack()
	return unregisterPlugin, deletePluginRoute, nil
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
			if err := s.pluginRouteOp.Create(ctx, tmpRoute); err != nil {
				return fmt.Errorf("error create tmpRoute(%s): %w", tmpRoute, err)
			}
			return nil
		})
	}
	return rbStack, nil
}
