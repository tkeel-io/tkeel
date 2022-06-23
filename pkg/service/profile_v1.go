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

	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel/api/profile/v1"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/plgprofile"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
)

type ProfileService struct {
	pb.UnimplementedProfileServer
	pluginOp    plugin.Operator
	ProfileOp   plgprofile.ProfileOperator
	daprHTTPCli dapr.Client
}

func NewProfileService(plgOp plugin.Operator, profileOp plgprofile.ProfileOperator, daprHTTP dapr.Client) *ProfileService {
	return &ProfileService{pluginOp: plgOp, ProfileOp: profileOp, daprHTTPCli: daprHTTP}
}

func (s *ProfileService) GetProfileSchema(ctx context.Context, request *pb.GetProfileSchemaRequest) (*pb.GetProfileSchemaResponse, error) {
	plugins, err := s.pluginOp.List(ctx)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrPluginList()
	}
	// ProfileOp set profileKey:plugin.
	for _, plg := range plugins {
		for k := range plg.Profiles {
			s.ProfileOp.SetProfilePlugin(ctx, k, plg.ID)
		}
	}

	profiles := make(map[string]*pb.ProfileSchema)
	for _, plg := range plugins {
		for k, prf := range plg.Profiles {
			profiles[k] = modelProfileSchema2pbProfile(prf)
		}
	}
	// keel profile.
	for keelP, keelV := range plgprofile.KeelProfiles {
		profiles[keelP] = modelProfileSchema2pbProfile(keelV)
	}

	return &pb.GetProfileSchemaResponse{Schema: &pb.Schema{
		Type:       "object",
		Properties: profiles,
	}}, nil
}

func (s *ProfileService) GetTenantProfileData(ctx context.Context, request *pb.GetTenantProfileDataRequest) (*pb.GetTenantProfileDataResponse, error) {
	data, err := s.ProfileOp.GetTenantProfileData(ctx, request.GetTenantId())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrInvalidArgument()
	}
	return &pb.GetTenantProfileDataResponse{Profiles: data}, nil
}

// nolint
func (s *ProfileService) SetTenantProfileData(ctx context.Context, request *pb.SetTenantPluginProfileRequest) (*pb.SetTenantPluginProfileResponse, error) {
	// set profile data.
	err := s.ProfileOp.SetTenantProfileData(ctx, request.GetTenantId(), request.GetBody().GetProfiles())
	if err != nil {
		log.Error(err)
		return nil, pb.ErrUnknown()
	}
	//  call plugin
	pluginProfiles := make(map[string]map[string]int64)
	for k, v := range request.GetBody().GetProfiles() {
		plg, _ := s.ProfileOp.GetProfilePlugin(ctx, k)
		if plg != "" && plg != plgprofile.PLUGIN_ID_KEEL {
			pluginProfiles[plg][k] = v
		}
	}
	for plg, profiles := range pluginProfiles {
		extrData := map[string]interface{}{"type": "SetProfile", "profiles": profiles}
		extrDataBytes, _ := json.Marshal(extrData)
		req, _ := json.Marshal(pb.TenantEnableRequest{TenantId: request.GetTenantId(), Extra: extrDataBytes})
		res, eCall := s.daprHTTPCli.Call(ctx, &dapr.AppRequest{
			ID:     plg,
			Method: "v1/tenant/enable",
			Verb:   "POST",
			Body:   req,
		})
		if eCall != nil {
			log.Error(err)
		}
		defer res.Body.Close()
	}
	return &pb.SetTenantPluginProfileResponse{}, nil
}

func modelProfileSchema2pbProfile(schema model.ProfileSchema) *pb.ProfileSchema {
	return &pb.ProfileSchema{Title: schema.Title,
		Type:        schema.Type,
		Default:     schema.Default,
		Description: schema.Description,
		MultipleOf:  schema.MultipleOf,
		Maximum:     schema.Maximum,
		Minimum:     schema.Minimum}
}
