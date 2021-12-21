// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// protoc-gen-go-http 0.1.0

package v1

import (
	context "context"
	go_restful "github.com/emicklei/go-restful"
	errors "github.com/tkeel-io/kit/errors"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	http "net/http"
)

import transportHTTP "github.com/tkeel-io/kit/transport/http"

// This is a compile-time assertion to ensure that this generated file
// is compatible with the tkeel package it is being compiled against.
// import package.context.http.go_restful.errors.emptypb.

type Oauth2HTTPServer interface {
	AddPluginWhiteList(context.Context, *AddPluginWhiteListRequest) (*emptypb.Empty, error)
	IssueAdminToken(context.Context, *IssueAdminTokenRequest) (*IssueTokenResponse, error)
	IssuePluginToken(context.Context, *IssuePluginTokenRequest) (*IssueTokenResponse, error)
}

type Oauth2HTTPHandler struct {
	srv Oauth2HTTPServer
}

func newOauth2HTTPHandler(s Oauth2HTTPServer) *Oauth2HTTPHandler {
	return &Oauth2HTTPHandler{srv: s}
}

func setResult(code int, msg string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}

func (h *Oauth2HTTPHandler) AddPluginWhiteList(req *go_restful.Request, resp *go_restful.Response) {
	in := AddPluginWhiteListRequest{}
	if err := transportHTTP.GetBody(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			setResult(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.AddPluginWhiteList(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			setResult(httpCode, tErr.Message, out), "application/json")
		return
	}

	resp.WriteHeaderAndJson(http.StatusOK,
		setResult(http.StatusOK, "ok", out), "application/json")
}

func (h *Oauth2HTTPHandler) IssueAdminToken(req *go_restful.Request, resp *go_restful.Response) {
	in := IssueAdminTokenRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			setResult(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.IssueAdminToken(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			setResult(httpCode, tErr.Message, out), "application/json")
		return
	}

	resp.WriteHeaderAndJson(http.StatusOK,
		setResult(http.StatusOK, "ok", out), "application/json")
}

func (h *Oauth2HTTPHandler) IssuePluginToken(req *go_restful.Request, resp *go_restful.Response) {
	in := IssuePluginTokenRequest{}
	if err := transportHTTP.GetBody(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			setResult(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.IssuePluginToken(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			setResult(httpCode, tErr.Message, out), "application/json")
		return
	}

	resp.WriteHeaderAndJson(http.StatusOK,
		setResult(http.StatusOK, "ok", out), "application/json")
}

func RegisterOauth2HTTPServer(container *go_restful.Container, srv Oauth2HTTPServer) {
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

	handler := newOauth2HTTPHandler(srv)
	ws.Route(ws.POST("/oauth2/plugin").
		To(handler.IssuePluginToken))
	ws.Route(ws.POST("/oauth2/plugin/white-list").
		To(handler.AddPluginWhiteList))
	ws.Route(ws.GET("/oauth2/admin").
		To(handler.IssueAdminToken))
}
