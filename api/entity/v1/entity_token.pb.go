// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.8.0
// source: api/entity/v1/entity_token.proto

package v1

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TokenRequestBody struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EntityType string `protobuf:"bytes,1,opt,name=entity_type,json=entityType,proto3" json:"entity_type,omitempty"`
	EntityId   string `protobuf:"bytes,2,opt,name=entity_id,json=entityId,proto3" json:"entity_id,omitempty"`
	Owner      string `protobuf:"bytes,3,opt,name=owner,proto3" json:"owner,omitempty"`
	ExpiresIn  int64  `protobuf:"varint,4,opt,name=expires_in,json=expiresIn,proto3" json:"expires_in,omitempty"`
}

func (x *TokenRequestBody) Reset() {
	*x = TokenRequestBody{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_entity_v1_entity_token_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TokenRequestBody) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenRequestBody) ProtoMessage() {}

func (x *TokenRequestBody) ProtoReflect() protoreflect.Message {
	mi := &file_api_entity_v1_entity_token_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenRequestBody.ProtoReflect.Descriptor instead.
func (*TokenRequestBody) Descriptor() ([]byte, []int) {
	return file_api_entity_v1_entity_token_proto_rawDescGZIP(), []int{0}
}

func (x *TokenRequestBody) GetEntityType() string {
	if x != nil {
		return x.EntityType
	}
	return ""
}

func (x *TokenRequestBody) GetEntityId() string {
	if x != nil {
		return x.EntityId
	}
	return ""
}

func (x *TokenRequestBody) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *TokenRequestBody) GetExpiresIn() int64 {
	if x != nil {
		return x.ExpiresIn
	}
	return 0
}

type CreateEntityTokenRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Body *TokenRequestBody `protobuf:"bytes,1,opt,name=body,proto3" json:"body,omitempty"`
}

func (x *CreateEntityTokenRequest) Reset() {
	*x = CreateEntityTokenRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_entity_v1_entity_token_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateEntityTokenRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateEntityTokenRequest) ProtoMessage() {}

func (x *CreateEntityTokenRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_entity_v1_entity_token_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateEntityTokenRequest.ProtoReflect.Descriptor instead.
func (*CreateEntityTokenRequest) Descriptor() ([]byte, []int) {
	return file_api_entity_v1_entity_token_proto_rawDescGZIP(), []int{1}
}

func (x *CreateEntityTokenRequest) GetBody() *TokenRequestBody {
	if x != nil {
		return x.Body
	}
	return nil
}

type CreateEntityTokenResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *CreateEntityTokenResponse) Reset() {
	*x = CreateEntityTokenResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_entity_v1_entity_token_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateEntityTokenResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateEntityTokenResponse) ProtoMessage() {}

func (x *CreateEntityTokenResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_entity_v1_entity_token_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateEntityTokenResponse.ProtoReflect.Descriptor instead.
func (*CreateEntityTokenResponse) Descriptor() ([]byte, []int) {
	return file_api_entity_v1_entity_token_proto_rawDescGZIP(), []int{2}
}

func (x *CreateEntityTokenResponse) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

type TokenInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *TokenInfoRequest) Reset() {
	*x = TokenInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_entity_v1_entity_token_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TokenInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenInfoRequest) ProtoMessage() {}

func (x *TokenInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_entity_v1_entity_token_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenInfoRequest.ProtoReflect.Descriptor instead.
func (*TokenInfoRequest) Descriptor() ([]byte, []int) {
	return file_api_entity_v1_entity_token_proto_rawDescGZIP(), []int{3}
}

func (x *TokenInfoRequest) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

type TokenInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EntityId   string `protobuf:"bytes,1,opt,name=entity_id,json=entityId,proto3" json:"entity_id,omitempty"`
	EntityType string `protobuf:"bytes,2,opt,name=entity_type,json=entityType,proto3" json:"entity_type,omitempty"`
	Owner      string `protobuf:"bytes,3,opt,name=owner,proto3" json:"owner,omitempty"`
	CreatedAt  int64  `protobuf:"varint,4,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	ExpiredAt  int64  `protobuf:"varint,5,opt,name=expired_at,json=expiredAt,proto3" json:"expired_at,omitempty"`
}

func (x *TokenInfoResponse) Reset() {
	*x = TokenInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_entity_v1_entity_token_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TokenInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TokenInfoResponse) ProtoMessage() {}

func (x *TokenInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_entity_v1_entity_token_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TokenInfoResponse.ProtoReflect.Descriptor instead.
func (*TokenInfoResponse) Descriptor() ([]byte, []int) {
	return file_api_entity_v1_entity_token_proto_rawDescGZIP(), []int{4}
}

func (x *TokenInfoResponse) GetEntityId() string {
	if x != nil {
		return x.EntityId
	}
	return ""
}

func (x *TokenInfoResponse) GetEntityType() string {
	if x != nil {
		return x.EntityType
	}
	return ""
}

func (x *TokenInfoResponse) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *TokenInfoResponse) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *TokenInfoResponse) GetExpiredAt() int64 {
	if x != nil {
		return x.ExpiredAt
	}
	return 0
}

var File_api_entity_v1_entity_token_proto protoreflect.FileDescriptor

var file_api_entity_v1_entity_token_proto_rawDesc = []byte{
	0x0a, 0x20, 0x61, 0x70, 0x69, 0x2f, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2f, 0x76, 0x31, 0x2f,
	0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0d, 0x61, 0x70, 0x69, 0x2e, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76,
	0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e,
	0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c,
	0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70,
	0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x98, 0x02,
	0x0a, 0x10, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x42, 0x6f,
	0x64, 0x79, 0x12, 0x41, 0x0a, 0x0b, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x20, 0x92, 0x41, 0x1a, 0x32, 0x18, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x20, 0x74, 0x79, 0x70, 0x65, 0x20, 0x20, 0x62, 0x6f, 0x64, 0x79, 0x20,
	0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0xe0, 0x41, 0x02, 0x52, 0x0a, 0x65, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x3b, 0x0a, 0x09, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1e, 0x92, 0x41, 0x18, 0x32, 0x16, 0x65,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x20, 0x69, 0x64, 0x20, 0x20, 0x62, 0x6f, 0x64, 0x79, 0x20, 0x70,
	0x61, 0x72, 0x61, 0x6d, 0x73, 0xe0, 0x41, 0x02, 0x52, 0x08, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x49, 0x64, 0x12, 0x37, 0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x21, 0x92, 0x41, 0x1b, 0x32, 0x19, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x20, 0x6f,
	0x77, 0x6e, 0x65, 0x72, 0x20, 0x20, 0x62, 0x6f, 0x64, 0x79, 0x20, 0x70, 0x61, 0x72, 0x61, 0x6d,
	0x73, 0xe0, 0x41, 0x02, 0x52, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x4b, 0x0a, 0x0a, 0x65,
	0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x5f, 0x69, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x42,
	0x2c, 0x92, 0x41, 0x26, 0x32, 0x24, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x20, 0x74, 0x6f, 0x6b,
	0x65, 0x6e, 0x20, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x20, 0x69, 0x6e, 0x20, 0x20, 0x62,
	0x6f, 0x64, 0x79, 0x20, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0xe0, 0x41, 0x02, 0x52, 0x09, 0x65,
	0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x49, 0x6e, 0x22, 0x79, 0x0a, 0x18, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x5d, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e,
	0x76, 0x31, 0x2e, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x42,
	0x6f, 0x64, 0x79, 0x42, 0x28, 0x92, 0x41, 0x22, 0x32, 0x20, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x20, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x20, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x20, 0x20, 0x62,
	0x6f, 0x64, 0x79, 0x20, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0xe0, 0x41, 0x02, 0x52, 0x04, 0x62,
	0x6f, 0x64, 0x79, 0x22, 0x31, 0x0a, 0x19, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x4b, 0x0a, 0x10, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x37, 0x0a, 0x05, 0x74, 0x6f,
	0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x21, 0x92, 0x41, 0x1b, 0x32, 0x19,
	0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x20, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x20, 0x20, 0x70, 0x61,
	0x74, 0x68, 0x20, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0xe0, 0x41, 0x02, 0x52, 0x05, 0x74, 0x6f,
	0x6b, 0x65, 0x6e, 0x22, 0xa5, 0x01, 0x0a, 0x11, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x1d, 0x0a,
	0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x1d, 0x0a, 0x0a,
	0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x09, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x64, 0x41, 0x74, 0x32, 0xc8, 0x04, 0x0a, 0x0b,
	0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0xcb, 0x01, 0x0a, 0x11,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x12, 0x27, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76,
	0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x54, 0x6f,
	0x6b, 0x65, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x63, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x15, 0x22, 0x0d, 0x2f, 0x65,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x2f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x3a, 0x04, 0x62, 0x6f, 0x64,
	0x79, 0x92, 0x41, 0x45, 0x0a, 0x0c, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x20, 0x74, 0x6f, 0x6b,
	0x65, 0x6e, 0x12, 0x15, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x20, 0x61, 0x20, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x20, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x2a, 0x11, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x4a, 0x0b, 0x0a, 0x03,
	0x32, 0x30, 0x30, 0x12, 0x04, 0x0a, 0x02, 0x4f, 0x4b, 0x12, 0xb4, 0x01, 0x0a, 0x09, 0x54, 0x6f,
	0x6b, 0x65, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1f, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x65,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x49, 0x6e,
	0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x64, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x16, 0x12, 0x14, 0x2f, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2f, 0x69, 0x6e, 0x66, 0x6f,
	0x2f, 0x7b, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x7d, 0x92, 0x41, 0x45, 0x0a, 0x0c, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x20, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x17, 0x67, 0x65, 0x74, 0x20, 0x61,
	0x20, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x20, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x20, 0x69, 0x6e,
	0x66, 0x6f, 0x2a, 0x0f, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x49,
	0x6e, 0x66, 0x6f, 0x4a, 0x0b, 0x0a, 0x03, 0x32, 0x30, 0x30, 0x12, 0x04, 0x0a, 0x02, 0x4f, 0x4b,
	0x12, 0xb3, 0x01, 0x0a, 0x11, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1f, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x65, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x17, 0x2a, 0x15, 0x2f, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x2f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x2f, 0x7b, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x7d, 0x92, 0x41,
	0x45, 0x0a, 0x0c, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x20, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12,
	0x15, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x20, 0x61, 0x20, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x20, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x2a, 0x11, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x45, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x4a, 0x0b, 0x0a, 0x03, 0x32, 0x30, 0x30,
	0x12, 0x04, 0x0a, 0x02, 0x4f, 0x4b, 0x42, 0x3d, 0x0a, 0x0d, 0x61, 0x70, 0x69, 0x2e, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x2e, 0x76, 0x31, 0x50, 0x01, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x6b, 0x65, 0x65, 0x6c, 0x2d, 0x69, 0x6f, 0x2f, 0x74,
	0x6b, 0x65, 0x65, 0x6c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2f,
	0x76, 0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_entity_v1_entity_token_proto_rawDescOnce sync.Once
	file_api_entity_v1_entity_token_proto_rawDescData = file_api_entity_v1_entity_token_proto_rawDesc
)

func file_api_entity_v1_entity_token_proto_rawDescGZIP() []byte {
	file_api_entity_v1_entity_token_proto_rawDescOnce.Do(func() {
		file_api_entity_v1_entity_token_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_entity_v1_entity_token_proto_rawDescData)
	})
	return file_api_entity_v1_entity_token_proto_rawDescData
}

var file_api_entity_v1_entity_token_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_api_entity_v1_entity_token_proto_goTypes = []interface{}{
	(*TokenRequestBody)(nil),          // 0: api.entity.v1.TokenRequestBody
	(*CreateEntityTokenRequest)(nil),  // 1: api.entity.v1.CreateEntityTokenRequest
	(*CreateEntityTokenResponse)(nil), // 2: api.entity.v1.CreateEntityTokenResponse
	(*TokenInfoRequest)(nil),          // 3: api.entity.v1.TokenInfoRequest
	(*TokenInfoResponse)(nil),         // 4: api.entity.v1.TokenInfoResponse
	(*emptypb.Empty)(nil),             // 5: google.protobuf.Empty
}
var file_api_entity_v1_entity_token_proto_depIdxs = []int32{
	0, // 0: api.entity.v1.CreateEntityTokenRequest.body:type_name -> api.entity.v1.TokenRequestBody
	1, // 1: api.entity.v1.EntityTokenOp.CreateEntityToken:input_type -> api.entity.v1.CreateEntityTokenRequest
	3, // 2: api.entity.v1.EntityTokenOp.TokenInfo:input_type -> api.entity.v1.TokenInfoRequest
	3, // 3: api.entity.v1.EntityTokenOp.DeleteEntityToken:input_type -> api.entity.v1.TokenInfoRequest
	2, // 4: api.entity.v1.EntityTokenOp.CreateEntityToken:output_type -> api.entity.v1.CreateEntityTokenResponse
	4, // 5: api.entity.v1.EntityTokenOp.TokenInfo:output_type -> api.entity.v1.TokenInfoResponse
	5, // 6: api.entity.v1.EntityTokenOp.DeleteEntityToken:output_type -> google.protobuf.Empty
	4, // [4:7] is the sub-list for method output_type
	1, // [1:4] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_entity_v1_entity_token_proto_init() }
func file_api_entity_v1_entity_token_proto_init() {
	if File_api_entity_v1_entity_token_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_entity_v1_entity_token_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TokenRequestBody); i {
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
		file_api_entity_v1_entity_token_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateEntityTokenRequest); i {
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
		file_api_entity_v1_entity_token_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateEntityTokenResponse); i {
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
		file_api_entity_v1_entity_token_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TokenInfoRequest); i {
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
		file_api_entity_v1_entity_token_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TokenInfoResponse); i {
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
			RawDescriptor: file_api_entity_v1_entity_token_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_entity_v1_entity_token_proto_goTypes,
		DependencyIndexes: file_api_entity_v1_entity_token_proto_depIdxs,
		MessageInfos:      file_api_entity_v1_entity_token_proto_msgTypes,
	}.Build()
	File_api_entity_v1_entity_token_proto = out.File
	file_api_entity_v1_entity_token_proto_rawDesc = nil
	file_api_entity_v1_entity_token_proto_goTypes = nil
	file_api_entity_v1_entity_token_proto_depIdxs = nil
}
