// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// protoc-gen-go-http 0.1.0

package v1

import (
	context "context"
	json "encoding/json"
	go_restful "github.com/emicklei/go-restful"
	errors "github.com/tkeel-io/kit/errors"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	http "net/http"
	reflect "reflect"
)

import transportHTTP "github.com/tkeel-io/kit/transport/http"

// This is a compile-time assertion to ensure that this generated file
// is compatible with the tkeel package it is being compiled against.
// import package.context.http.reflect.go_restful.json.errors.emptypb.

type EntryHTTPServer interface {
	GetEntries(context.Context, *emptypb.Empty) (*GetEntriesResponse, error)
}

type EntryHTTPHandler struct {
	srv EntryHTTPServer
}

func newEntryHTTPHandler(s EntryHTTPServer) *EntryHTTPHandler {
	return &EntryHTTPHandler{srv: s}
}

func (h *EntryHTTPHandler) GetEntries(req *go_restful.Request, resp *go_restful.Response) {
	in := emptypb.Empty{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.GetEntries(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteErrorString(httpCode, tErr.Message)
		return
	}
	if reflect.ValueOf(out).Elem().Type().AssignableTo(reflect.TypeOf(emptypb.Empty{})) {
		resp.WriteHeader(http.StatusNoContent)
		return
	}
	result, err := json.Marshal(out)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	_, err = resp.Write(result)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
}

func RegisterEntryHTTPServer(container *go_restful.Container, srv EntryHTTPServer) {
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

	handler := newEntryHTTPHandler(srv)
	ws.Route(ws.GET("/entries").
		To(handler.GetEntries))
}
