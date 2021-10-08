package keel

import (
	"context"
	"math/rand"
	"net/http"

	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/logger"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/plugin"
	"github.com/tkeel-io/tkeel/pkg/utils"
)

var (
	log = logger.NewLogger("keel.plugin.keel")
)

type Keel struct {
	p *plugin.Plugin
}

func New(p *plugin.Plugin) (*Keel, error) {
	return &Keel{
		p: p,
	}, nil
}

func (m *Keel) Run() {
	pID := m.p.Conf().Plugin.ID
	if pID == "" {
		log.Fatal("error plugin id: \"\"")
	}
	if pID != "keel" {
		log.Fatalf("error plugin id: %s should be keel", pID)
	}

	go func() {
		m.p.SetRequiredFunc(openapi.RequiredFunc{
			Identify: func() (*openapi.IdentifyResp, error) {
				return &openapi.IdentifyResp{
					CommonResult: openapi.SuccessResult(),
					PluginID:     m.p.GetIdentifyResp().PluginID,
					Version:      m.p.GetIdentifyResp().Version,
					AddonsPoints: []*openapi.AddonsPoint{
						{
							AddonsPoint: "externalPreRouteCheck",
							Desc: `
							callback before external flow routing
							input request header and path
							output http statecode
							200   -- allow
							other -- deny
							`,
						},
					},
				}, nil
			},
		})
		m.p.SetOptionalFunc(openapi.OptionalFunc{
			AddonsIdentify: func(air *openapi.AddonsIdentifyReq) (*openapi.AddonsIdentifyResp, error) {
				for _, v := range air.Endpoint {
					xKeel := rand.Int() % 2
					xKeelStr := "False"
					if xKeel == 1 {
						xKeelStr = "True"
					}
					resp, err := keel.CallKeel(context.TODO(), air.Plugin.ID, v.Endpoint,
						http.MethodGet, &keel.CallReq{
							Header: http.Header{
								"x-keel-check": []string{xKeelStr},
							},
							Body: []byte(`Check whether the endpoint correctly implements this callback:
						For example, when the request header contains the "x-keel-check" field, 
						the HTTP request header 200 is returned. When the field value is "True", 
						the body is 
						{	
							"msg":"ok",
							"ret":0
						}, 
						When it is False, the body is 
						{
							"msg":"faild",
							"ret":-1
						}. 
						If it is not included, it will judge whether the request is valid.`),
						})
					if err != nil {
						return &openapi.AddonsIdentifyResp{
							CommonResult: openapi.BadRequestResult(resp.Status),
						}, nil
					}
					result := &openapi.CommonResult{}
					if err := utils.ReadBody2Json(resp.Body, result); err != nil {
						log.Errorf("error read addons identify(%s/%s/%s) resp: %s",
							air.Plugin.ID, v.Endpoint, v.AddonsPoint, err.Error())
						return &openapi.AddonsIdentifyResp{
							CommonResult: openapi.BadRequestResult(err.Error()),
						}, nil
					}
					if xKeel == 1 {
						if result.Ret != 0 || result.Msg != "ok" {
							log.Errorf("error identify(%s/%s/%s) resp: %v",
								air.Plugin.ID, v.Endpoint, v.AddonsPoint, result)
							return &openapi.AddonsIdentifyResp{
								CommonResult: openapi.BadRequestResult(resp.Status),
							}, nil
						}
					} else {
						if result.Ret != -1 || result.Msg != "faild" {
							log.Errorf("error identify(%s/%s/%s) resp: %v",
								air.Plugin.ID, v.Endpoint, v.AddonsPoint, result)
							return &openapi.AddonsIdentifyResp{
								CommonResult: openapi.BadRequestResult(resp.Status),
							}, nil
						}
					}
				}
				return &openapi.AddonsIdentifyResp{
					CommonResult: openapi.SuccessResult(),
				}, nil
			},
		})
		err := m.p.Run([]*openapi.API{
			{

				Endpoint: "/",
				H:        m.Route,
			},
		}...)
		if err != nil {
			log.Fatalf("error plugin run: %s", err)
			return
		}
	}()
	log.Debug("keel runing")
}
