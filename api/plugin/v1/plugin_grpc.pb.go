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

// PluginClient is the client API for Plugin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PluginClient interface {
	InstallPlugin(ctx context.Context, in *InstallPluginRequest, opts ...grpc.CallOption) (*InstallPluginResponse, error)
	UninstallPlugin(ctx context.Context, in *UninstallPluginRequest, opts ...grpc.CallOption) (*UninstallPluginResponse, error)
	GetPlugin(ctx context.Context, in *GetPluginRequest, opts ...grpc.CallOption) (*GetPluginResponse, error)
	ListPlugin(ctx context.Context, in *ListPluginRequest, opts ...grpc.CallOption) (*ListPluginResponse, error)
	TenantEnable(ctx context.Context, in *TenantEnableRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	TenantDisable(ctx context.Context, in *TenantDisableRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ListEnabledTenants(ctx context.Context, in *ListEnabledTenantsRequest, opts ...grpc.CallOption) (*ListEnabledTenantsResponse, error)
	TMUpdatePluginIdentify(ctx context.Context, in *TMUpdatePluginIdentifyRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	TMRegisterPlugin(ctx context.Context, in *TMRegisterPluginRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type pluginClient struct {
	cc grpc.ClientConnInterface
}

func NewPluginClient(cc grpc.ClientConnInterface) PluginClient {
	return &pluginClient{cc}
}

func (c *pluginClient) InstallPlugin(ctx context.Context, in *InstallPluginRequest, opts ...grpc.CallOption) (*InstallPluginResponse, error) {
	out := new(InstallPluginResponse)
	err := c.cc.Invoke(ctx, "/io.tkeel.rudder.api.plugin.v1.Plugin/InstallPlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) UninstallPlugin(ctx context.Context, in *UninstallPluginRequest, opts ...grpc.CallOption) (*UninstallPluginResponse, error) {
	out := new(UninstallPluginResponse)
	err := c.cc.Invoke(ctx, "/io.tkeel.rudder.api.plugin.v1.Plugin/UninstallPlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) GetPlugin(ctx context.Context, in *GetPluginRequest, opts ...grpc.CallOption) (*GetPluginResponse, error) {
	out := new(GetPluginResponse)
	err := c.cc.Invoke(ctx, "/io.tkeel.rudder.api.plugin.v1.Plugin/GetPlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) ListPlugin(ctx context.Context, in *ListPluginRequest, opts ...grpc.CallOption) (*ListPluginResponse, error) {
	out := new(ListPluginResponse)
	err := c.cc.Invoke(ctx, "/io.tkeel.rudder.api.plugin.v1.Plugin/ListPlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) TenantEnable(ctx context.Context, in *TenantEnableRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/io.tkeel.rudder.api.plugin.v1.Plugin/TenantEnable", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) TenantDisable(ctx context.Context, in *TenantDisableRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/io.tkeel.rudder.api.plugin.v1.Plugin/TenantDisable", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) ListEnabledTenants(ctx context.Context, in *ListEnabledTenantsRequest, opts ...grpc.CallOption) (*ListEnabledTenantsResponse, error) {
	out := new(ListEnabledTenantsResponse)
	err := c.cc.Invoke(ctx, "/io.tkeel.rudder.api.plugin.v1.Plugin/ListEnabledTenants", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) TMUpdatePluginIdentify(ctx context.Context, in *TMUpdatePluginIdentifyRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/io.tkeel.rudder.api.plugin.v1.Plugin/TMUpdatePluginIdentify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) TMRegisterPlugin(ctx context.Context, in *TMRegisterPluginRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/io.tkeel.rudder.api.plugin.v1.Plugin/TMRegisterPlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PluginServer is the server API for Plugin service.
// All implementations must embed UnimplementedPluginServer
// for forward compatibility
type PluginServer interface {
	InstallPlugin(context.Context, *InstallPluginRequest) (*InstallPluginResponse, error)
	UninstallPlugin(context.Context, *UninstallPluginRequest) (*UninstallPluginResponse, error)
	GetPlugin(context.Context, *GetPluginRequest) (*GetPluginResponse, error)
	ListPlugin(context.Context, *ListPluginRequest) (*ListPluginResponse, error)
	TenantEnable(context.Context, *TenantEnableRequest) (*emptypb.Empty, error)
	TenantDisable(context.Context, *TenantDisableRequest) (*emptypb.Empty, error)
	ListEnabledTenants(context.Context, *ListEnabledTenantsRequest) (*ListEnabledTenantsResponse, error)
	TMUpdatePluginIdentify(context.Context, *TMUpdatePluginIdentifyRequest) (*emptypb.Empty, error)
	TMRegisterPlugin(context.Context, *TMRegisterPluginRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedPluginServer()
}

// UnimplementedPluginServer must be embedded to have forward compatible implementations.
type UnimplementedPluginServer struct {
}

func (UnimplementedPluginServer) InstallPlugin(context.Context, *InstallPluginRequest) (*InstallPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InstallPlugin not implemented")
}
func (UnimplementedPluginServer) UninstallPlugin(context.Context, *UninstallPluginRequest) (*UninstallPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UninstallPlugin not implemented")
}
func (UnimplementedPluginServer) GetPlugin(context.Context, *GetPluginRequest) (*GetPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPlugin not implemented")
}
func (UnimplementedPluginServer) ListPlugin(context.Context, *ListPluginRequest) (*ListPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPlugin not implemented")
}
func (UnimplementedPluginServer) TenantEnable(context.Context, *TenantEnableRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TenantEnable not implemented")
}
func (UnimplementedPluginServer) TenantDisable(context.Context, *TenantDisableRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TenantDisable not implemented")
}
func (UnimplementedPluginServer) ListEnabledTenants(context.Context, *ListEnabledTenantsRequest) (*ListEnabledTenantsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListEnabledTenants not implemented")
}
func (UnimplementedPluginServer) TMUpdatePluginIdentify(context.Context, *TMUpdatePluginIdentifyRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TMUpdatePluginIdentify not implemented")
}
func (UnimplementedPluginServer) TMRegisterPlugin(context.Context, *TMRegisterPluginRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TMRegisterPlugin not implemented")
}
func (UnimplementedPluginServer) mustEmbedUnimplementedPluginServer() {}

// UnsafePluginServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PluginServer will
// result in compilation errors.
type UnsafePluginServer interface {
	mustEmbedUnimplementedPluginServer()
}

func RegisterPluginServer(s grpc.ServiceRegistrar, srv PluginServer) {
	s.RegisterService(&Plugin_ServiceDesc, srv)
}

func _Plugin_InstallPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InstallPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).InstallPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.tkeel.rudder.api.plugin.v1.Plugin/InstallPlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).InstallPlugin(ctx, req.(*InstallPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_UninstallPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UninstallPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).UninstallPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.tkeel.rudder.api.plugin.v1.Plugin/UninstallPlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).UninstallPlugin(ctx, req.(*UninstallPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_GetPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).GetPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.tkeel.rudder.api.plugin.v1.Plugin/GetPlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).GetPlugin(ctx, req.(*GetPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_ListPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).ListPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.tkeel.rudder.api.plugin.v1.Plugin/ListPlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).ListPlugin(ctx, req.(*ListPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_TenantEnable_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TenantEnableRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).TenantEnable(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.tkeel.rudder.api.plugin.v1.Plugin/TenantEnable",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).TenantEnable(ctx, req.(*TenantEnableRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_TenantDisable_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TenantDisableRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).TenantDisable(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.tkeel.rudder.api.plugin.v1.Plugin/TenantDisable",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).TenantDisable(ctx, req.(*TenantDisableRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_ListEnabledTenants_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListEnabledTenantsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).ListEnabledTenants(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.tkeel.rudder.api.plugin.v1.Plugin/ListEnabledTenants",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).ListEnabledTenants(ctx, req.(*ListEnabledTenantsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_TMUpdatePluginIdentify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TMUpdatePluginIdentifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).TMUpdatePluginIdentify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.tkeel.rudder.api.plugin.v1.Plugin/TMUpdatePluginIdentify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).TMUpdatePluginIdentify(ctx, req.(*TMUpdatePluginIdentifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_TMRegisterPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TMRegisterPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).TMRegisterPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.tkeel.rudder.api.plugin.v1.Plugin/TMRegisterPlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).TMRegisterPlugin(ctx, req.(*TMRegisterPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Plugin_ServiceDesc is the grpc.ServiceDesc for Plugin service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Plugin_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "io.tkeel.rudder.api.plugin.v1.Plugin",
	HandlerType: (*PluginServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InstallPlugin",
			Handler:    _Plugin_InstallPlugin_Handler,
		},
		{
			MethodName: "UninstallPlugin",
			Handler:    _Plugin_UninstallPlugin_Handler,
		},
		{
			MethodName: "GetPlugin",
			Handler:    _Plugin_GetPlugin_Handler,
		},
		{
			MethodName: "ListPlugin",
			Handler:    _Plugin_ListPlugin_Handler,
		},
		{
			MethodName: "TenantEnable",
			Handler:    _Plugin_TenantEnable_Handler,
		},
		{
			MethodName: "TenantDisable",
			Handler:    _Plugin_TenantDisable_Handler,
		},
		{
			MethodName: "ListEnabledTenants",
			Handler:    _Plugin_ListEnabledTenants_Handler,
		},
		{
			MethodName: "TMUpdatePluginIdentify",
			Handler:    _Plugin_TMUpdatePluginIdentify_Handler,
		},
		{
			MethodName: "TMRegisterPlugin",
			Handler:    _Plugin_TMRegisterPlugin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/plugin/v1/plugin.proto",
}
