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
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest, err.Error(), nil), "application/json")
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.GetEntries(ctx, &in)
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
