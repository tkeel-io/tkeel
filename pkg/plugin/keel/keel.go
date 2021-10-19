package keel

import (
	"context"
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/logger"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/plugin"
	"github.com/tkeel-io/tkeel/pkg/utils"
)

var log = logger.NewLogger("keel.plugin.keel")

type Keel struct {
	p *plugin.Plugin
}

func New(p *plugin.Plugin) (*Keel, error) {
	return &Keel{
		p: p,
	}, nil
}

func (k *Keel) Run() {
	pID := k.p.Conf().Plugin.ID
	if pID == "" {
		log.Fatal("error plugin id: \"\"")
	}
	if pID != "keel" {
		log.Fatalf("error plugin id: %s should be keel", pID)
	}

	go func() {
		k.p.SetRequiredFunc(openapi.RequiredFunc{
			Identify: k.identify,
		})
		k.p.SetOptionalFunc(openapi.OptionalFunc{
			AddonsIdentify: k.addonsIdentify,
		})
		err := k.p.Run(&openapi.API{Endpoint: "/", H: k.Route})
		if err != nil {
			log.Fatalf("error plugin run: %s", err)
			return
		}
	}()
	log.Debug("keel running")
}

func (k *Keel) identify() (*openapi.IdentifyResp, error) {
	return &openapi.IdentifyResp{
		CommonResult: openapi.SuccessResult(),
		PluginID:     k.p.GetIdentifyResp().PluginID,
		Version:      k.p.GetIdentifyResp().Version,
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
}

func (k *Keel) addonsIdentify(air *openapi.AddonsIdentifyReq) (*openapi.AddonsIdentifyResp, error) {
	endpointReq := air.Endpoint[0]
	xKeelStr := getRandBoolStr()

	resp, err := keel.CallKeel(context.TODO(), air.Plugin.ID, endpointReq.Endpoint,
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
		log.Errorf("error addons identify: %w", err)
		return &openapi.AddonsIdentifyResp{
			CommonResult: openapi.BadRequestResult(resp.Status),
		}, nil
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Errorf("error response body close: %s", err)
		}
	}()
	result := &openapi.CommonResult{}
	if err := utils.ReadBody2Json(resp.Body, result); err != nil {
		log.Errorf("error read addons identify(%s/%s/%s) resp: %s",
			air.Plugin.ID, endpointReq.Endpoint, endpointReq.AddonsPoint, err.Error())
		return &openapi.AddonsIdentifyResp{
			CommonResult: openapi.BadRequestResult(err.Error()),
		}, nil
	}
	if (xKeelStr == "True" && result.Ret == 0 && result.Msg == "ok") ||
		(xKeelStr == "False" && result.Ret == -1 && result.Msg == "faild") {
		return &openapi.AddonsIdentifyResp{
			CommonResult: openapi.SuccessResult(),
		}, nil
	}
	log.Errorf("error addons check identify(%s/%s/%s) resp: %v",
		air.Plugin.ID, endpointReq.Endpoint, endpointReq.AddonsPoint, result)
	return &openapi.AddonsIdentifyResp{
		CommonResult: openapi.BadRequestResult(resp.Status),
	}, nil
}

func getRandBoolStr() string {
	n, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		log.Errorf("error rand: %w", err)
		return "False"
	}
	if n.Int64()%2 == 1 {
		return "True"
	}
	return "False"
}
