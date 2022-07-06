package v1

import (
	"context"
	"net/http"

	go_restful "github.com/emicklei/go-restful"
	"github.com/tkeel-io/kit/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/kit/result"
	"github.com/tkeel-io/tkeel/pkg/model"
)

type NotificationHTTPServer interface {
	GetNotification(context.Context, string) (interface{}, error)
}

func RegisterNotificationHTTPServer(container *go_restful.Container, srv NotificationHTTPServer) {
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

	handler := newNotificationHTTPHandler(srv)
	ws.Route(ws.GET("/notifications").
		To(handler.GetNotification))
}

func newNotificationHTTPHandler(s NotificationHTTPServer) *NotificationHTTPHandler {
	return &NotificationHTTPHandler{srv: s}
}

type NotificationHTTPHandler struct {
	srv NotificationHTTPServer
}

func (n NotificationHTTPHandler) GetNotification(req *go_restful.Request, resp *go_restful.Response) {
	xTKAUth := req.Request.Header.Get(model.XtKeelAuthHeader)
	user := new(model.User)
	err := user.Base64Decode(xTKAUth)
	log.Info(user)
	if xTKAUth == "" || err != nil {
		tErr := errors.FromError(EntryErrInvalidTenant())
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		if httpCode == http.StatusMovedPermanently {
			resp.Header().Set("Location", tErr.Message)
		}
		resp.WriteHeaderAndJson(httpCode,
			result.Set(tErr.Reason, tErr.Message, nil), "application/json")
		return
	}
	notifications, _ := n.srv.GetNotification(req.Request.Context(), user.Tenant)
	resultData := map[string]interface{}{"code": errors.Success.Reason,
		"msg":  "",
		"data": map[string]interface{}{"notifications": notifications},
	}

	resp.WriteHeaderAndJson(200, resultData, "application/json")
}
