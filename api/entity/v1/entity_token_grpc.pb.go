// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// EntityTokenClient is the client API for EntityTokenOp service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EntityTokenClient interface {
	CreateEntityToken(ctx context.Context, in *CreateEntityTokenRequest, opts ...grpc.CallOption) (*CreateEntityTokenResponse, error)
	TokenInfo(ctx context.Context, in *TokenInfoRequest, opts ...grpc.CallOption) (*TokenInfoResponse, error)
	DeleteEntityToken(ctx context.Context, in *TokenInfoRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type entityTokenClient struct {
	cc grpc.ClientConnInterface
}

func NewEntityTokenClient(cc grpc.ClientConnInterface) EntityTokenClient {
	return &entityTokenClient{cc}
}

func (c *entityTokenClient) CreateEntityToken(ctx context.Context, in *CreateEntityTokenRequest, opts ...grpc.CallOption) (*CreateEntityTokenResponse, error) {
	out := new(CreateEntityTokenResponse)
	err := c.cc.Invoke(ctx, "/api.entity.v1.EntityTokenOp/CreateEntityToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *entityTokenClient) TokenInfo(ctx context.Context, in *TokenInfoRequest, opts ...grpc.CallOption) (*TokenInfoResponse, error) {
	out := new(TokenInfoResponse)
	err := c.cc.Invoke(ctx, "/api.entity.v1.EntityTokenOp/TokenInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *entityTokenClient) DeleteEntityToken(ctx context.Context, in *TokenInfoRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/api.entity.v1.EntityTokenOp/DeleteEntityToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EntityTokenServer is the server API for EntityTokenOp service.
// All implementations must embed UnimplementedEntityTokenServer
// for forward compatibility
type EntityTokenServer interface {
	CreateEntityToken(context.Context, *CreateEntityTokenRequest) (*CreateEntityTokenResponse, error)
	TokenInfo(context.Context, *TokenInfoRequest) (*TokenInfoResponse, error)
	DeleteEntityToken(context.Context, *TokenInfoRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedEntityTokenServer()
}

// UnimplementedEntityTokenServer must be embedded to have forward compatible implementations.
type UnimplementedEntityTokenServer struct {
}

func (UnimplementedEntityTokenServer) CreateEntityToken(context.Context, *CreateEntityTokenRequest) (*CreateEntityTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateEntityToken not implemented")
}
func (UnimplementedEntityTokenServer) TokenInfo(context.Context, *TokenInfoRequest) (*TokenInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TokenInfo not implemented")
}
func (UnimplementedEntityTokenServer) DeleteEntityToken(context.Context, *TokenInfoRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteEntityToken not implemented")
}
func (UnimplementedEntityTokenServer) mustEmbedUnimplementedEntityTokenServer() {}

// UnsafeEntityTokenServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EntityTokenServer will
// result in compilation errors.
type UnsafeEntityTokenServer interface {
	mustEmbedUnimplementedEntityTokenServer()
}

func RegisterEntityTokenServer(s grpc.ServiceRegistrar, srv EntityTokenServer) {
	s.RegisterService(&EntityToken_ServiceDesc, srv)
}

func _EntityToken_CreateEntityToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateEntityTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EntityTokenServer).CreateEntityToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.entity.v1.EntityTokenOp/CreateEntityToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EntityTokenServer).CreateEntityToken(ctx, req.(*CreateEntityTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EntityToken_TokenInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TokenInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EntityTokenServer).TokenInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.entity.v1.EntityTokenOp/TokenInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EntityTokenServer).TokenInfo(ctx, req.(*TokenInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EntityToken_DeleteEntityToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TokenInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EntityTokenServer).DeleteEntityToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.entity.v1.EntityTokenOp/DeleteEntityToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EntityTokenServer).DeleteEntityToken(ctx, req.(*TokenInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// EntityToken_ServiceDesc is the grpc.ServiceDesc for EntityTokenOp service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EntityToken_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.entity.v1.EntityTokenOp",
	HandlerType: (*EntityTokenServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateEntityToken",
			Handler:    _EntityToken_CreateEntityToken_Handler,
		},
		{
			MethodName: "TokenInfo",
			Handler:    _EntityToken_TokenInfo_Handler,
		},
		{
			MethodName: "DeleteEntityToken",
			Handler:    _EntityToken_DeleteEntityToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/entity/v1/entity_token.proto",
}
