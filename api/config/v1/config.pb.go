//
//Copyright 2021 The tKeel Authors.
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: api/config/v1/config.proto

package v1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetDeploymentConfigResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AdminHost  string `protobuf:"bytes,1,opt,name=admin_host,json=adminHost,proto3" json:"admin_host,omitempty"`
	TenantHost string `protobuf:"bytes,2,opt,name=tenant_host,json=tenantHost,proto3" json:"tenant_host,omitempty"`
	Port       string `protobuf:"bytes,3,opt,name=port,proto3" json:"port,omitempty"`
	DocsAddr   string `protobuf:"bytes,4,opt,name=docs_addr,json=docsAddr,proto3" json:"docs_addr,omitempty"`
}

func (x *GetDeploymentConfigResponse) Reset() {
	*x = GetDeploymentConfigResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_config_v1_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDeploymentConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDeploymentConfigResponse) ProtoMessage() {}

func (x *GetDeploymentConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_config_v1_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDeploymentConfigResponse.ProtoReflect.Descriptor instead.
func (*GetDeploymentConfigResponse) Descriptor() ([]byte, []int) {
	return file_api_config_v1_config_proto_rawDescGZIP(), []int{0}
}

func (x *GetDeploymentConfigResponse) GetAdminHost() string {
	if x != nil {
		return x.AdminHost
	}
	return ""
}

func (x *GetDeploymentConfigResponse) GetTenantHost() string {
	if x != nil {
		return x.TenantHost
	}
	return ""
}

func (x *GetDeploymentConfigResponse) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

func (x *GetDeploymentConfigResponse) GetDocsAddr() string {
	if x != nil {
		return x.DocsAddr
	}
	return ""
}

type PlatformConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key  string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Path string `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
}

func (x *PlatformConfigRequest) Reset() {
	*x = PlatformConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_config_v1_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PlatformConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlatformConfigRequest) ProtoMessage() {}

func (x *PlatformConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_config_v1_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlatformConfigRequest.ProtoReflect.Descriptor instead.
func (*PlatformConfigRequest) Descriptor() ([]byte, []int) {
	return file_api_config_v1_config_proto_rawDescGZIP(), []int{1}
}

func (x *PlatformConfigRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *PlatformConfigRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type SetPlatformExtraConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string          `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Path  string          `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Extra *structpb.Value `protobuf:"bytes,3,opt,name=extra,proto3" json:"extra,omitempty"`
}

func (x *SetPlatformExtraConfigRequest) Reset() {
	*x = SetPlatformExtraConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_config_v1_config_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetPlatformExtraConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetPlatformExtraConfigRequest) ProtoMessage() {}

func (x *SetPlatformExtraConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_config_v1_config_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetPlatformExtraConfigRequest.ProtoReflect.Descriptor instead.
func (*SetPlatformExtraConfigRequest) Descriptor() ([]byte, []int) {
	return file_api_config_v1_config_proto_rawDescGZIP(), []int{2}
}

func (x *SetPlatformExtraConfigRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *SetPlatformExtraConfigRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *SetPlatformExtraConfigRequest) GetExtra() *structpb.Value {
	if x != nil {
		return x.Extra
	}
	return nil
}

var File_api_config_v1_config_proto protoreflect.FileDescriptor

var file_api_config_v1_config_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1d, 0x69, 0x6f,
	0x2e, 0x74, 0x6b, 0x65, 0x65, 0x6c, 0x2e, 0x72, 0x75, 0x64, 0x64, 0x65, 0x72, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e,
	0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xda, 0x01, 0x0a, 0x1b, 0x47, 0x65, 0x74, 0x44, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x33, 0x0a, 0x0a, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x5f, 0x68, 0x6f,
	0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x14, 0x92, 0x41, 0x11, 0x32, 0x0f, 0xe7,
	0xae, 0xa1, 0xe7, 0x90, 0x86, 0xe7, 0xab, 0xaf, 0xe5, 0x9f, 0x9f, 0xe5, 0x90, 0x8d, 0x52, 0x09,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x35, 0x0a, 0x0b, 0x74, 0x65, 0x6e,
	0x61, 0x6e, 0x74, 0x5f, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x14,
	0x92, 0x41, 0x11, 0x32, 0x0f, 0xe7, 0xa7, 0x9f, 0xe6, 0x88, 0xb7, 0xe7, 0xab, 0xaf, 0xe5, 0x9f,
	0x9f, 0xe5, 0x90, 0x8d, 0x52, 0x0a, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x48, 0x6f, 0x73, 0x74,
	0x12, 0x1f, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x0b,
	0x92, 0x41, 0x08, 0x32, 0x06, 0xe7, 0xab, 0xaf, 0xe5, 0x8f, 0xa3, 0x52, 0x04, 0x70, 0x6f, 0x72,
	0x74, 0x12, 0x2e, 0x0a, 0x09, 0x64, 0x6f, 0x63, 0x73, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x11, 0x92, 0x41, 0x0e, 0x32, 0x0c, 0xe6, 0x96, 0x87, 0xe6, 0xa1,
	0xa3, 0xe5, 0x9c, 0xb0, 0xe5, 0x9d, 0x80, 0x52, 0x08, 0x64, 0x6f, 0x63, 0x73, 0x41, 0x64, 0x64,
	0x72, 0x22, 0x52, 0x0a, 0x15, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0x92, 0x41, 0x05, 0x32, 0x03, 0x6b, 0x65,
	0x79, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x1d, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x09, 0x92, 0x41, 0x06, 0x32, 0x04, 0x70, 0x61, 0x74, 0x68, 0x52,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x22, 0x94, 0x01, 0x0a, 0x1d, 0x53, 0x65, 0x74, 0x50, 0x6c, 0x61,
	0x74, 0x66, 0x6f, 0x72, 0x6d, 0x45, 0x78, 0x74, 0x72, 0x61, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0x92, 0x41, 0x05, 0x32, 0x03, 0x6b, 0x65, 0x79, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x1d, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x09, 0x92, 0x41, 0x06, 0x32, 0x04, 0x70, 0x61, 0x74, 0x68, 0x52, 0x04, 0x70, 0x61,
	0x74, 0x68, 0x12, 0x38, 0x0a, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x0a, 0x92, 0x41, 0x07, 0x32, 0x05,
	0x65, 0x78, 0x74, 0x72, 0x61, 0x52, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x32, 0xb6, 0x08, 0x0a,
	0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x86, 0x02, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x44,
	0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12,
	0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x3a, 0x2e, 0x69, 0x6f, 0x2e, 0x74, 0x6b, 0x65,
	0x65, 0x6c, 0x2e, 0x72, 0x75, 0x64, 0x64, 0x65, 0x72, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x65, 0x70, 0x6c, 0x6f,
	0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x9a, 0x01, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x14, 0x12, 0x12, 0x2f, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74,
	0x92, 0x41, 0x7d, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1f, 0xe6, 0x9f, 0xa5,
	0xe8, 0xaf, 0xa2, 0x20, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x20, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x20, 0xe6, 0x8e, 0xa5, 0xe5, 0x8f, 0xa3, 0x2a, 0x13, 0x47, 0x65,
	0x74, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x4a, 0x0b, 0x0a, 0x03, 0x32, 0x30, 0x30, 0x12, 0x04, 0x0a, 0x02, 0x4f, 0x4b, 0x4a, 0x17,
	0x0a, 0x03, 0x34, 0x30, 0x30, 0x12, 0x10, 0x0a, 0x0e, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44,
	0x5f, 0x54, 0x45, 0x4e, 0x41, 0x4e, 0x54, 0x4a, 0x17, 0x0a, 0x03, 0x35, 0x30, 0x30, 0x12, 0x10,
	0x0a, 0x0e, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x4e, 0x41, 0x4c, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52,
	0x12, 0xf8, 0x01, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x34, 0x2e, 0x69, 0x6f, 0x2e, 0x74, 0x6b, 0x65, 0x65,
	0x6c, 0x2e, 0x72, 0x75, 0x64, 0x64, 0x65, 0x72, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x22, 0x94, 0x01, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x12, 0x10, 0x2f,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x92,
	0x41, 0x79, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1d, 0xe6, 0x9f, 0xa5, 0xe8,
	0xaf, 0xa2, 0x20, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x20, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x20, 0xe6, 0x8e, 0xa5, 0xe5, 0x8f, 0xa3, 0x2a, 0x11, 0x47, 0x65, 0x74, 0x50, 0x6c,
	0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x4a, 0x0b, 0x0a, 0x03,
	0x32, 0x30, 0x30, 0x12, 0x04, 0x0a, 0x02, 0x4f, 0x4b, 0x4a, 0x17, 0x0a, 0x03, 0x34, 0x30, 0x30,
	0x12, 0x10, 0x0a, 0x0e, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x54, 0x45, 0x4e, 0x41,
	0x4e, 0x54, 0x4a, 0x17, 0x0a, 0x03, 0x35, 0x30, 0x30, 0x12, 0x10, 0x0a, 0x0e, 0x49, 0x4e, 0x54,
	0x45, 0x52, 0x4e, 0x41, 0x4c, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x12, 0xff, 0x01, 0x0a, 0x11,
	0x44, 0x65, 0x6c, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0x34, 0x2e, 0x69, 0x6f, 0x2e, 0x74, 0x6b, 0x65, 0x65, 0x6c, 0x2e, 0x72, 0x75, 0x64,
	0x64, 0x65, 0x72, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x76,
	0x31, 0x2e, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22,
	0x9b, 0x01, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x19, 0x2a, 0x17, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2f, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2f, 0x75, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x92, 0x41, 0x79, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1d, 0xe5, 0x88,
	0xa0, 0xe9, 0x99, 0xa4, 0x20, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x20, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x20, 0xe6, 0x8e, 0xa5, 0xe5, 0x8f, 0xa3, 0x2a, 0x11, 0x44, 0x65, 0x6c,
	0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x4a, 0x0b,
	0x0a, 0x03, 0x32, 0x30, 0x30, 0x12, 0x04, 0x0a, 0x02, 0x4f, 0x4b, 0x4a, 0x17, 0x0a, 0x03, 0x34,
	0x30, 0x30, 0x12, 0x10, 0x0a, 0x0e, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x54, 0x45,
	0x4e, 0x41, 0x4e, 0x54, 0x4a, 0x17, 0x0a, 0x03, 0x35, 0x30, 0x30, 0x12, 0x10, 0x0a, 0x0e, 0x49,
	0x4e, 0x54, 0x45, 0x52, 0x4e, 0x41, 0x4c, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x12, 0xa5, 0x02,
	0x0a, 0x16, 0x53, 0x65, 0x74, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x45, 0x78, 0x74,
	0x72, 0x61, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x3c, 0x2e, 0x69, 0x6f, 0x2e, 0x74, 0x6b,
	0x65, 0x65, 0x6c, 0x2e, 0x72, 0x75, 0x64, 0x64, 0x65, 0x72, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x74, 0x50, 0x6c, 0x61, 0x74,
	0x66, 0x6f, 0x72, 0x6d, 0x45, 0x78, 0x74, 0x72, 0x61, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0xb4,
	0x01, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x20, 0x22, 0x17, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x2f, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x3a, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x92, 0x41, 0x8a, 0x01, 0x0a, 0x06, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x12, 0x29, 0xe8, 0xae, 0xbe, 0xe7, 0xbd, 0xae, 0x20, 0x70, 0x6c, 0x61, 0x74,
	0x66, 0x6f, 0x72, 0x6d, 0x20, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x20, 0x65, 0x78, 0x74, 0x72,
	0x61, 0x20, 0xe6, 0x95, 0xb0, 0xe6, 0x8d, 0xae, 0xe6, 0x8e, 0xa5, 0xe5, 0x8f, 0xa3, 0x2a, 0x16,
	0x53, 0x65, 0x74, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x45, 0x78, 0x74, 0x72, 0x61,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x4a, 0x0b, 0x0a, 0x03, 0x32, 0x30, 0x30, 0x12, 0x04, 0x0a,
	0x02, 0x4f, 0x4b, 0x4a, 0x17, 0x0a, 0x03, 0x34, 0x30, 0x30, 0x12, 0x10, 0x0a, 0x0e, 0x49, 0x4e,
	0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x54, 0x45, 0x4e, 0x41, 0x4e, 0x54, 0x4a, 0x17, 0x0a, 0x03,
	0x35, 0x30, 0x30, 0x12, 0x10, 0x0a, 0x0e, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x4e, 0x41, 0x4c, 0x5f,
	0x45, 0x52, 0x52, 0x4f, 0x52, 0x42, 0x4d, 0x0a, 0x1d, 0x69, 0x6f, 0x2e, 0x74, 0x6b, 0x65, 0x65,
	0x6c, 0x2e, 0x72, 0x75, 0x64, 0x64, 0x65, 0x72, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x2e, 0x76, 0x31, 0x50, 0x01, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x6b, 0x65, 0x65, 0x6c, 0x2d, 0x69, 0x6f, 0x2f, 0x74, 0x6b,
	0x65, 0x65, 0x6c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x76,
	0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_config_v1_config_proto_rawDescOnce sync.Once
	file_api_config_v1_config_proto_rawDescData = file_api_config_v1_config_proto_rawDesc
)

func file_api_config_v1_config_proto_rawDescGZIP() []byte {
	file_api_config_v1_config_proto_rawDescOnce.Do(func() {
		file_api_config_v1_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_config_v1_config_proto_rawDescData)
	})
	return file_api_config_v1_config_proto_rawDescData
}

var file_api_config_v1_config_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_api_config_v1_config_proto_goTypes = []interface{}{
	(*GetDeploymentConfigResponse)(nil),   // 0: io.tkeel.rudder.api.config.v1.GetDeploymentConfigResponse
	(*PlatformConfigRequest)(nil),         // 1: io.tkeel.rudder.api.config.v1.PlatformConfigRequest
	(*SetPlatformExtraConfigRequest)(nil), // 2: io.tkeel.rudder.api.config.v1.SetPlatformExtraConfigRequest
	(*structpb.Value)(nil),                // 3: google.protobuf.Value
	(*emptypb.Empty)(nil),                 // 4: google.protobuf.Empty
}
var file_api_config_v1_config_proto_depIdxs = []int32{
	3, // 0: io.tkeel.rudder.api.config.v1.SetPlatformExtraConfigRequest.extra:type_name -> google.protobuf.Value
	4, // 1: io.tkeel.rudder.api.config.v1.Config.GetDeploymentConfig:input_type -> google.protobuf.Empty
	1, // 2: io.tkeel.rudder.api.config.v1.Config.GetPlatformConfig:input_type -> io.tkeel.rudder.api.config.v1.PlatformConfigRequest
	1, // 3: io.tkeel.rudder.api.config.v1.Config.DelPlatformConfig:input_type -> io.tkeel.rudder.api.config.v1.PlatformConfigRequest
	2, // 4: io.tkeel.rudder.api.config.v1.Config.SetPlatformExtraConfig:input_type -> io.tkeel.rudder.api.config.v1.SetPlatformExtraConfigRequest
	0, // 5: io.tkeel.rudder.api.config.v1.Config.GetDeploymentConfig:output_type -> io.tkeel.rudder.api.config.v1.GetDeploymentConfigResponse
	3, // 6: io.tkeel.rudder.api.config.v1.Config.GetPlatformConfig:output_type -> google.protobuf.Value
	3, // 7: io.tkeel.rudder.api.config.v1.Config.DelPlatformConfig:output_type -> google.protobuf.Value
	3, // 8: io.tkeel.rudder.api.config.v1.Config.SetPlatformExtraConfig:output_type -> google.protobuf.Value
	5, // [5:9] is the sub-list for method output_type
	1, // [1:5] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_config_v1_config_proto_init() }
func file_api_config_v1_config_proto_init() {
	if File_api_config_v1_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_config_v1_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDeploymentConfigResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_config_v1_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PlatformConfigRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_config_v1_config_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetPlatformExtraConfigRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_config_v1_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_config_v1_config_proto_goTypes,
		DependencyIndexes: file_api_config_v1_config_proto_depIdxs,
		MessageInfos:      file_api_config_v1_config_proto_msgTypes,
	}.Build()
	File_api_config_v1_config_proto = out.File
	file_api_config_v1_config_proto_rawDesc = nil
	file_api_config_v1_config_proto_goTypes = nil
	file_api_config_v1_config_proto_depIdxs = nil
}
