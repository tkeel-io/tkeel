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
	"gorm.io/gorm"
	"sync"

	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel/api/profile/v1"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/client/openapi"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/metrics"
	"github.com/tkeel-io/tkeel/pkg/model/plgprofile"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
)

type ProfileService struct {
	pb.UnimplementedProfileServer
	pluginOp       plugin.Operator
	ProfileOp      plgprofile.ProfileOperator
	daprHTTPCli    dapr.Client
	openapiClient  openapi.Client
	defaultProfile map[string]int32
	tenantDB       *gorm.DB
}

func NewProfileService(plgOp plugin.Operator, profileOp plgprofile.ProfileOperator, daprHTTP dapr.Client, openapiClient openapi.Client, tenantDB *gorm.DB) *ProfileService {
	return &ProfileService{pluginOp: plgOp, ProfileOp: profileOp, daprHTTPCli: daprHTTP, openapiClient: openapiClient, defaultProfile: make(map[string]int32), tenantDB: tenantDB}
}

func (s *ProfileService) GetProfileSchema(ctx context.Context, request *pb.GetProfileSchemaRequest) (*pb.GetProfileSchemaResponse, error) {
	plugins, err := s.pluginOp.List(ctx)
	if err != nil {
		log.Error(err)
		return nil, pb.ErrPluginList()
	}
	profiles := make(map[string]*pb.ProfileSchema)
	required := make([]string, 0)
	// ProfileOp set profileKey:plugin.
	for _, plg := range plugins {
		// call identify.
		identify, errIdf := s.openapiClient.Identify(ctx, plg.ID)
		if errIdf != nil {
			log.Error(errIdf)
			continue
		}
		for k, prf := range identify.Profiles {
			s.ProfileOp.SetProfilePlugin(ctx, k, plg.ID)
			profiles[k] = &pb.ProfileSchema{Title: prf.Title, Type: prf.Type, Default: prf.Default, Description: prf.Description, MultipleOf: prf.MultipleOf, Maximum: prf.Maximum, Minimum: prf.Minimum}
			s.defaultProfile[k] = prf.Default
			required = append(required, k)
		}
	}

	// keel profile.
	for keelP, keelV := range plgprofile.KeelProfiles {
		profiles[keelP] = modelProfileSchema2pbProfile(keelV)
		s.defaultProfile[keelP] = keelV.Default
		required = append(required, keelP)
	}

	return &pb.GetProfileSchemaResponse{Schema: &pb.Schema{
		Type:                 "object",
		Properties:           profiles,
		Required:             required,
		AdditionalProperties: false,
	}}, nil
}

func (s *ProfileService) GetTenantProfileData(ctx context.Context, request *pb.GetTenantProfileDataRequest) (*pb.GetTenantProfileDataResponse, error) {
	data, _ := s.ProfileOp.GetTenantProfileData(ctx, request.GetTenantId())
	if data == nil {
		data = s.defaultProfile
	}
	for k, v := range data {
		metrics.CollectorTKeelProfiles.WithLabelValues(request.GetTenantId(), k).Set(float64(v))
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
	pluginProfiles := make(map[string]map[string]int32)
	for k, v := range request.GetBody().GetProfiles() {
		plg, _ := s.ProfileOp.GetProfilePlugin(ctx, k)
		if plg != "" && plg != plgprofile.PLUGIN_ID_KEEL {
			if pluginProfiles[plg] == nil {
				pluginProfiles[plg] = make(map[string]int32)
			}
			pluginProfiles[plg][k] = v
		}
	}
	var wg sync.WaitGroup
	for plg, profiles := range pluginProfiles {
		wg.Add(1)
		go func(plugin string, profile map[string]int32) {
			extrData := map[string]interface{}{"type": "SetProfile", "profiles": profile}
			extrDataBytes, _ := json.Marshal(extrData)
			req, _ := json.Marshal(pb.TenantEnableRequest{TenantId: request.GetTenantId(), Extra: extrDataBytes})
			res, eCall := s.daprHTTPCli.Call(ctx, &dapr.AppRequest{
				ID:     plugin,
				Method: "v1/tenant/enable",
				Verb:   "POST",
				Body:   req,
			})
			if eCall != nil {
				log.Error(err)
			}
			for pk, pv := range profile {
				metrics.CollectorTKeelProfiles.WithLabelValues(request.GetTenantId(), pk).Set(float64(pv))
			}
			defer res.Body.Close()
			wg.Done()
		}(plg, profiles)
	}
	wg.Wait()
	return &pb.SetTenantPluginProfileResponse{}, nil
}

func modelProfileSchema2pbProfile(schema *model.ProfileSchema) *pb.ProfileSchema {
	return &pb.ProfileSchema{Title: schema.Title,
		Type:        schema.Type,
		Default:     schema.Default,
		Description: schema.Description,
		MultipleOf:  schema.MultipleOf,
		Maximum:     schema.Maximum,
		Minimum:     schema.Minimum}
}
