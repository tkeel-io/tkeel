// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// protoc-gen-go-http 0.1.0

package v1

import (
	context "context"
	go_restful "github.com/emicklei/go-restful"
	errors "github.com/tkeel-io/kit/errors"
	result "github.com/tkeel-io/kit/result"
	protojson "google.golang.org/protobuf/encoding/protojson"
	anypb "google.golang.org/protobuf/types/known/anypb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	http "net/http"
)

import transportHTTP "github.com/tkeel-io/kit/transport/http"

// This is a compile-time assertion to ensure that this generated file
// is compatible with the tkeel package it is being compiled against.
// import package.context.http.anypb.result.protojson.go_restful.errors.emptypb.

var (
	_ = protojson.MarshalOptions{}
	_ = anypb.Any{}
	_ = emptypb.Empty{}
)

type TenantHTTPServer interface {
	AddTenantPlugin(context.Context, *AddTenantPluginRequest) (*AddTenantPluginResponse, error)
	CreateTenant(context.Context, *CreateTenantRequest) (*CreateTenantResponse, error)
	CreateUser(context.Context, *CreateUserRequest) (*CreateUserResponse, error)
	DeleteTenant(context.Context, *DeleteTenantRequest) (*emptypb.Empty, error)
	DeleteTenantPlugin(context.Context, *DeleteTenantPluginRequest) (*DeleteTenantPluginResponse, error)
	DeleteUser(context.Context, *DeleteUserRequest) (*emptypb.Empty, error)
	GetTenant(context.Context, *GetTenantRequest) (*GetTenantResponse, error)
	GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
	ListTenant(context.Context, *emptypb.Empty) (*ListTenantResponse, error)
	ListTenantPlugin(context.Context, *ListTenantPluginRequest) (*ListTenantPluginResponse, error)
	ListUser(context.Context, *ListUserRequest) (*ListUserResponse, error)
	TenantPluginPermissible(context.Context, *PluginPermissibleRequest) (*PluginPermissibleResponse, error)
	UpdateUser(context.Context, *UpdateUserRequest) (*UpdateUserResponse, error)
}

type TenantHTTPHandler struct {
	srv TenantHTTPServer
}

func newTenantHTTPHandler(s TenantHTTPServer) *TenantHTTPHandler {
	return &TenantHTTPHandler{srv: s}
}

func (h *TenantHTTPHandler) AddTenantPlugin(req *go_restful.Request, resp *go_restful.Response) {
	in := AddTenantPluginRequest{}
	if err := transportHTTP.GetBody(req, &in.Body); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.AddTenantPlugin(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) CreateTenant(req *go_restful.Request, resp *go_restful.Response) {
	in := CreateTenantRequest{}
	if err := transportHTTP.GetBody(req, &in.Body); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.CreateTenant(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) CreateUser(req *go_restful.Request, resp *go_restful.Response) {
	in := CreateUserRequest{}
	if err := transportHTTP.GetBody(req, &in.Body); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.CreateUser(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) DeleteTenant(req *go_restful.Request, resp *go_restful.Response) {
	in := DeleteTenantRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.DeleteTenant(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) DeleteTenantPlugin(req *go_restful.Request, resp *go_restful.Response) {
	in := DeleteTenantPluginRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.DeleteTenantPlugin(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) DeleteUser(req *go_restful.Request, resp *go_restful.Response) {
	in := DeleteUserRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.DeleteUser(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) GetTenant(req *go_restful.Request, resp *go_restful.Response) {
	in := GetTenantRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.GetTenant(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) GetUser(req *go_restful.Request, resp *go_restful.Response) {
	in := GetUserRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.GetUser(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) ListTenant(req *go_restful.Request, resp *go_restful.Response) {
	in := emptypb.Empty{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.ListTenant(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) ListTenantPlugin(req *go_restful.Request, resp *go_restful.Response) {
	in := ListTenantPluginRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.ListTenantPlugin(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) ListUser(req *go_restful.Request, resp *go_restful.Response) {
	in := ListUserRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.ListUser(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) TenantPluginPermissible(req *go_restful.Request, resp *go_restful.Response) {
	in := PluginPermissibleRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.TenantPluginPermissible(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func (h *TenantHTTPHandler) UpdateUser(req *go_restful.Request, resp *go_restful.Response) {
	in := UpdateUserRequest{}
	if err := transportHTTP.GetBody(req, &in.Body); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.UpdateUser(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg:  "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
}

func RegisterTenantHTTPServer(container *go_restful.Container, srv TenantHTTPServer) {
	var ws *go_restful.WebService
	for _, v := range container.RegisteredWebServices() {
		if v.RootPath() == "/v1" {
			ws = v
			break
		}
	}
	if ws == nil {
		ws = new(go_restful.WebService)
		ws.ApiVersion("/v1")
		ws.Path("/v1").Produces(go_restful.MIME_JSON)
		container.Add(ws)
	}

	handler := newTenantHTTPHandler(srv)
	ws.Route(ws.POST("/tenants").
		To(handler.CreateTenant))
	ws.Route(ws.GET("/tenants/{tenant_id}").
		To(handler.GetTenant))
	ws.Route(ws.GET("/tenants").
		To(handler.ListTenant))
	ws.Route(ws.DELETE("/tenants/{tenant_id}").
		To(handler.DeleteTenant))
	ws.Route(ws.POST("/tenants/{tenant_id}/users").
		To(handler.CreateUser))
	ws.Route(ws.GET("/tenants/{tenant_id}/users/{user_id}").
		To(handler.GetUser))
	ws.Route(ws.GET("/tenants/{tenant_id}/users").
		To(handler.ListUser))
	ws.Route(ws.DELETE("/tenants/{tenant_id}/users/{user_id}").
		To(handler.DeleteUser))
	ws.Route(ws.PUT("/tenants/{tenant_id}/users/{user_id}").
		To(handler.UpdateUser))
	ws.Route(ws.POST("/tenants/{tenant_id}/plugins").
		To(handler.AddTenantPlugin))
	ws.Route(ws.GET("/tenants/{tenant_id}/plugins").
		To(handler.ListTenantPlugin))
	ws.Route(ws.DELETE("/tenants/{tenant_id}/plugins/{plugin_id}").
		To(handler.DeleteTenantPlugin))
	ws.Route(ws.GET("/tenants/plugins/permissible").
		To(handler.TenantPluginPermissible))
}
