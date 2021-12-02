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
	"errors"
	"fmt"

	"github.com/tkeel-io/kit/log"
	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	pb "github.com/tkeel-io/tkeel/api/plugin/v1"
	"github.com/tkeel-io/tkeel/pkg/client/openapi"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	"github.com/tkeel-io/tkeel/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrGetOpenapiIdentify = errors.New("error get openapi identify")
	ErrPluginRegistered   = errors.New("plugin is registered")
)

type PluginServiceV1 struct {
	pb.UnimplementedPluginServer

	tkeelConf     *config.TkeelConf
	pluginOp      plugin.Operator
	pluginRouteOp proute.Operator
	openapiClient openapi.Client
}

func NewPluginServiceV1(conf *config.TkeelConf, pluginOperator plugin.Operator,
	prouteOperator proute.Operator, openapi openapi.Client) *PluginServiceV1 {
	return &PluginServiceV1{
		tkeelConf:     conf,
		pluginOp:      pluginOperator,
		pluginRouteOp: prouteOperator,
		openapiClient: openapi,
	}
}

func (s *PluginServiceV1) RegisterPlugin(ctx context.Context,
	req *pb.RegisterPluginRequest) (retResp *emptypb.Empty, err error) {
	// get register plugin identify.
	resp, err := s.queryIdentify(ctx, req)
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
	if err = s.registerPlugin(ctx, req, resp); err != nil {
		log.Errorf("error register plugin: %s", err)
		if errors.Is(err, ErrPluginRegistered) {
			return nil, pb.PluginErrPluginAlreadyExists()
		}
		return nil, pb.PluginErrInternalQueryPluginOpenapi()
	}
	log.Debugf("register plugin(%s) ok", req.Id)
	return &emptypb.Empty{}, nil
}

func (s *PluginServiceV1) DeletePlugin(ctx context.Context,
	req *pb.DeletePluginRequest) (*pb.DeletePluginResponse, error) {
	deleteID := req.Id
	// check exists.
	deletePluginRoute, err := s.pluginRouteOp.Get(ctx, deleteID)
	if err != nil {
		log.Errorf("error delete plugin(%s) route get: %s", deleteID, err)
		if errors.Is(err, proute.ErrPluginRouteNotExsist) {
			return nil, pb.PluginErrPluginNotFound()
		}
		return nil, pb.PluginErrInternalStore()
	}
	// check whether the extension point is implemented.
	if len(deletePluginRoute.RegisterAddons) != 0 {
		log.Errorf("error delete plugin(%s): other plugins have implemented the extension points of this plugin.", deleteID)
		return nil, pb.PluginErrDeletePluginHasBeenDepended()
	}

	// delete plugin.
	delPlugin, delPluginRoute, err := s.deletePlugin(ctx, deleteID)
	if err != nil {
		log.Errorf("error delete plugin(%s): %s", deleteID, err)
		return nil, pb.PluginErrInternalStore()
	}
	return &pb.DeletePluginResponse{
		Plugin: util.ConvertModel2PluginObjectPb(delPlugin, delPluginRoute),
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
		pr, err := s.pluginRouteOp.Get(ctx, p.ID)
		if err != nil {
			log.Errorf("error plugin list get plugin(%s) route: %s", p.ID, err)
			return nil, pb.PluginErrInternalStore()
		}
		retList = append(retList, util.ConvertModel2PluginObjectPb(p, pr))
	}

	return &pb.ListPluginResponse{
		PluginList: retList,
	}, nil
}

func (s *PluginServiceV1) queryIdentify(ctx context.Context,
	req *pb.RegisterPluginRequest) (*openapi_v1.IdentifyResponse, error) {
	pID := req.Id
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
		return fmt.Errorf("error plugin(%s) depend tkeel version(%s) not invaild",
			resp.PluginId, resp.TkeelVersion)
	}
	return nil
}

func (s *PluginServiceV1) registerPlugin(ctx context.Context, req *pb.RegisterPluginRequest, resp *openapi_v1.IdentifyResponse) (err error) {
	rbStack := util.NewRollbackStack()
	defer func() {
		if err != nil {
			rbStack.Run()
		}
	}()
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
	// save new plugin and update plugin route.
	err = s.saveNewPlugin(ctx, resp, req.Secret, tmpPluginRoute)
	if err != nil {
		return fmt.Errorf("error save new plugin(%s): %w", resp.PluginId, err)
	}
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
	resp *openapi_v1.IdentifyResponse) (retList []util.RollbackFunc, err error) {
	rbList := util.NewRollbackStack()
	defer func() {
		if err != nil {
			rbList.Run()
		}
	}()
	for _, v := range resp.ImplementedPlugin {
		oldPluginRoute, err := s.pluginRouteOp.Get(ctx, v.Plugin.Id)
		if err != nil {
			return nil, fmt.Errorf("error implemented plugin(%s) not registered", v.Plugin.Id)
		}
		ok, err := util.CheckRegisterPluginTkeelVersion(oldPluginRoute.TkeelVersion, resp.TkeelVersion)
		if err != nil {
			return nil, fmt.Errorf("error check implemented plugin(%s) depened tkeel version: %w",
				v.Plugin.Id, err)
		}
		if !ok {
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
			return nil, fmt.Errorf("error addons identify(%s/%s): %w", v.Plugin.Id, addonsReq, err)
		}
		if addonsResp.Res.Ret != openapi_v1.Retcode_OK {
			return nil, fmt.Errorf("error addons identify(%s/%s): %s", v.Plugin.Id, addonsReq, addonsResp.Res.Msg)
		}
		pluginRouteBackup := oldPluginRoute.Clone()
		util.UpdatePluginRoute(resp.PluginId, v.Addons, oldPluginRoute)
		rbList = append(rbList, func() error {
			log.Debugf("register implemented plugin route roll back run.")
			err = s.pluginRouteOp.Update(ctx, pluginRouteBackup)
			if err != nil {
				return fmt.Errorf("error update plugin route backup(%s): %w", pluginRouteBackup, err)
			}
			return nil
		})
		err = s.pluginRouteOp.Update(ctx, oldPluginRoute)
		if err != nil {
			return nil, fmt.Errorf("error update old plugin route(%s): %w", oldPluginRoute, err)
		}
	}
	return rbList, nil
}

func (s *PluginServiceV1) saveNewPlugin(ctx context.Context, resp *openapi_v1.IdentifyResponse,
	secret string, tmpPluginRoute *model.PluginRoute) (err error) {
	rbList := util.NewRollbackStack()
	defer func() {
		if err != nil {
			rbList.Run()
		}
	}()
	statusResp, err := s.openapiClient.Status(ctx, resp.PluginId)
	if err != nil {
		return fmt.Errorf("error status(%s): %w", resp.PluginId, err)
	}
	if statusResp.Res.Ret != openapi_v1.Retcode_OK {
		return fmt.Errorf("error status(%s): %s", resp.PluginId, statusResp.Res.Msg)
	}
	newPlugin := model.NewPlugin(resp, secret)
	err = s.pluginOp.Create(ctx, newPlugin)
	if err != nil {
		return fmt.Errorf("error create plugin(%s): %w", newPlugin, err)
	}
	rbList = append(rbList, func() error {
		log.Debugf("save new plugin roll back run.")
		_, err = s.pluginOp.Delete(ctx, newPlugin.ID)
		if err != nil {
			return fmt.Errorf("error delete newPlugin(%s): %w", newPlugin.ID, err)
		}
		return nil
	})
	tmpPluginRoute.Status = statusResp.Status
	err = s.pluginRouteOp.Update(ctx, tmpPluginRoute)
	if err != nil {
		return fmt.Errorf("error update tmp plugin route(%s): %w", tmpPluginRoute, err)
	}
	return nil
}

func (s *PluginServiceV1) deletePlugin(ctx context.Context, deleteID string) (p *model.Plugin, pr *model.PluginRoute, err error) {
	rbStack := util.NewRollbackStack()
	defer func() {
		if err != nil {
			rbStack.Run()
		}
	}()
	deletePlugin, err := s.pluginOp.Delete(ctx, deleteID)
	if err != nil {
		return nil, nil, fmt.Errorf("error delete plugin(%s): %w", deleteID, err)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("delete plugin roll back run.")
		err = s.pluginOp.Create(ctx, deletePlugin)
		if err != nil {
			return fmt.Errorf("error create delete plugin(%s): %w", deletePlugin, err)
		}
		return nil
	})
	// delete plugin route.
	deletePluginRoute, err := s.pluginRouteOp.Delete(ctx, deleteID)
	if err != nil {
		return nil, nil, fmt.Errorf("error delete plugin(%s) route: %w", deleteID, err)
	}
	rbStack = append(rbStack, func() error {
		log.Debugf("delete plugin route roll back run.")
		err = s.pluginRouteOp.Create(ctx, deletePluginRoute)
		if err != nil {
			return fmt.Errorf("error create delete plugin(%s) route: %w", deletePluginRoute, err)
		}
		return nil
	})
	// reset implemented plugin route.
	_, err = s.resetImplementedPluginRoute(ctx, deletePlugin)
	if err != nil {
		return nil, nil, fmt.Errorf("error reset implemented plugin route(%s): %w", deleteID, err)
	}
	return deletePlugin, deletePluginRoute, nil
}

func (s *PluginServiceV1) resetImplementedPluginRoute(ctx context.Context,
	deletePlugin *model.Plugin) (retStack []util.RollbackFunc, err error) {
	rbStack := util.NewRollbackStack()
	defer func() {
		if err != nil {
			rbStack.Run()
		}
	}()
	for _, v := range deletePlugin.ImplementedPlugin {
		oldRoute, err := s.pluginRouteOp.Get(ctx, v.Plugin.Id)
		if err != nil {
			return nil, fmt.Errorf("error plugin route(%s) get: %w", v.Plugin.Id, err)
		}
		tmpRoute := oldRoute.Clone()
		for _, a := range v.Addons {
			delete(oldRoute.RegisterAddons, a.AddonsPoint)
		}
		err = s.pluginRouteOp.Update(ctx, oldRoute)
		if err != nil {
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
