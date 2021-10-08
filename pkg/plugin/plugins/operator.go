package plugins

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/utils"
)

func requestPluginIdentify(ctx context.Context, pID string) (*openapi.IdentifyResp, error) {
	identifyResp := &openapi.IdentifyResp{}
	resp, err := keel.CallPlugin(ctx, pID,
		openapi.IDENTIFY_METHOD, http.MethodGet,
		nil)
	if err != nil {
		log.Errorf("error get plugin(%s) identify: %s", pID, err)
		return nil, err
	}
	defer resp.Body.Close()

	err = utils.ReadBody2Json(resp.Body, identifyResp)
	if err != nil {
		log.Errorf("error get plugin(%s) identify read body: %s", pID, err)
		return nil, err
	}
	if identifyResp.Ret != 0 {
		err = fmt.Errorf("error plugin(%s) get identify: %s",
			identifyResp.PluginID, identifyResp.Msg)
		log.Error(err)
		return nil, err
	}
	// check plugin id matched
	if identifyResp.PluginID != pID {
		err = fmt.Errorf("error plugin(%s) identify not match(%s)",
			pID, identifyResp.PluginID)
		log.Error(err)
		return nil, err
	}
	return identifyResp, nil
}

func requestPluginAddonsIdentify(ctx context.Context, pID string,
	mp *openapi.MainPlugin) (*openapi.AddonsIdentifyResp, error) {
	// check addons vaild
	addonsIdentifyRep := &openapi.AddonsIdentifyReq{
		Plugin: struct {
			ID      string "json:\"id\""
			Version string "json:\"version\""
		}{
			ID:      pID,
			Version: mp.Version,
		},
		Endpoint: mp.Endpoints,
	}
	respByte, err := json.Marshal(addonsIdentifyRep)
	if err != nil {
		log.Errorf("error marshal addons resp(%s) : %s",
			mp.ID, err)
		return nil, err
	}
	addonsIdentifyResp := &openapi.AddonsIdentifyResp{}
	resp, err := keel.CallPlugin(ctx, mp.ID,
		openapi.ADDONSIDENTIFY_METHOD, http.MethodPost,
		&keel.CallReq{
			Body: respByte,
		})
	if err != nil {
		log.Errorf("error post plugin(%s) addons identify: %s",
			mp.ID, err)
		return nil, err
	}
	defer resp.Body.Close()

	err = utils.ReadBody2Json(resp.Body, addonsIdentifyResp)
	if err != nil {
		log.Errorf("error post plugin(%s) read body: %s",
			pID, err)
		return nil, err
	}
	if addonsIdentifyResp.Ret != 0 {
		err = fmt.Errorf("error plugin(%s) check addons identify(%s): %s",
			pID, mp.ID, addonsIdentifyResp.Msg)
		log.Error(err)
		return nil, err
	}
	return addonsIdentifyResp, nil
}

func requestPluginTenantBind(ctx context.Context, pID string,
	tenantID string, extra []byte) (*openapi.TenantBindResp, error) {
	// check addons vaild
	tenantBindRep := &openapi.TenantBindReq{
		TenantID: tenantID,
		Extra:    extra,
	}
	reqByte, err := json.Marshal(tenantBindRep)
	if err != nil {
		return nil, fmt.Errorf("error marshal tenant bind request: %s", err)
	}
	tenantBindResp := &openapi.TenantBindResp{}
	resp, err := keel.CallPlugin(ctx, pID,
		openapi.TENANTBIND_METHOD, http.MethodPost,
		&keel.CallReq{
			Body: reqByte,
		})
	if err != nil {
		log.Errorf("error post plugin(%s) tenant bind: %s",
			pID, err)
		return nil, err
	}
	defer resp.Body.Close()

	err = utils.ReadBody2Json(resp.Body, tenantBindResp)
	if err != nil {
		log.Errorf("error post plugin(%s) read body: %s",
			pID, err)
		return nil, err
	}
	if tenantBindResp.Ret != 0 {
		err = fmt.Errorf("error plugin(%s) request plugin tenant bind(%s): %s",
			pID, tenantID, tenantBindResp.Msg)
		log.Error(err)
		return nil, err
	}
	return tenantBindResp, nil
}

func requestPluginStatus(ctx context.Context, pID string) (*openapi.StatusResp, error) {
	// TODO 在最后进行存储，而不是现在这样，插件回调就进行存储
	statusResp := &openapi.StatusResp{}
	resp, err := keel.CallPlugin(ctx, pID, openapi.STATUS_METHOD,
		http.MethodGet, nil)
	if err != nil {
		log.Errorf("error get plugin(%s) status: %s", pID, err)
		return nil, err
	}
	defer resp.Body.Close()

	err = utils.ReadBody2Json(resp.Body, statusResp)
	if err != nil {
		log.Errorf("error get plugin(%s) status read body: %s", pID, err)
		return nil, err
	}
	if statusResp.Ret != 0 {
		err = fmt.Errorf("error plugin(%s) get status: %s",
			pID, statusResp.Msg)
		log.Error(err)
		return nil, err
	}
	return statusResp, nil
}

func scrapePluginStatus(ctx context.Context, interval time.Duration) error {
	// Concurrency lock
	f, etag, err := keel.GetScrapeFlag(ctx)
	if err != nil {
		return err
	}
	if f != "" || etag != "" {
		log.Debug("failed to grab lock")
		return nil
	}
	err = keel.SaveScrapeFlag(ctx, etag, interval.Milliseconds()/1000)
	if err != nil {
		return err
	}

	log.Debug("successfully to grab lock")
	allPluginMap, allEtag, err := keel.GetAllRegisteredPlugin(ctx)
	if err != nil {
		return err
	}
	for pID := range allPluginMap {
		pRoute, pREtag, err := keel.GetPluginRoute(ctx, pID)
		if err != nil {
			log.Errorf("error get plugin(%s) route: %s when scrape plugin status", pID, err)
			continue
		}
		statusResp, err := requestPluginStatus(ctx, pID)
		if err != nil {
			log.Errorf("error request plugin(%s) status: %s when scrape plugin status", pID, err)
			continue
		}
		allPluginMap[pID] = string(statusResp.Status)
		if pRoute.Status != statusResp.Status {
			pRoute.Status = statusResp.Status
			if pREtag == "" {
				pREtag = "-1"
			}
			err := keel.SavePluginRoute(ctx, pID, pRoute, pREtag)
			if err != nil {
				log.Errorf("error save plugin(%s) route: %s when scrape plugin status", pID, err)
				continue
			}
		}
	}
	err = keel.SaveAllRegisteredPlugin(ctx, allPluginMap, allEtag)
	if err != nil {
		return err
	}
	return nil
}

func registerPlugin(ctx context.Context, identifyResp *openapi.IdentifyResp, Secret string) (err error) {
	var pID string

	if identifyResp != nil {
		pID = identifyResp.PluginID
	}
	// check if plugin has been registered
	_, etag, err := keel.GetPlugin(ctx, pID)
	if err != nil {
		return err
	}
	if etag != "" {
		log.Debugf("error plugin(%s) has been registered", pID)
		return fmt.Errorf("plugin(%s) has been registered", pID)
	}

	pRoute, prEtag, err := keel.GetPluginRoute(ctx, pID)
	if err != nil {
		return err
	}
	if prEtag != "" {
		log.Debugf("error plugin(%s) has been registered or be registering(%s)", pID, pRoute.Status)
		return fmt.Errorf("plugin(%s) has been registered or be registering(%s)", pID, pRoute.Status)
	}

	err = keel.SavePluginRoute(ctx, pID, &keel.PluginRoute{
		Status: openapi.Starting,
	}, "-1")
	if err != nil {
		return fmt.Errorf("error save plugin route: %s", err)
	}
	defer func() {
		if err != nil {
			keel.DeletePluginRoute(ctx, pID)
		}
	}()

	newPlugin := &keel.Plugin{
		IdentifyResp: identifyResp,
		Secret:       Secret,
		RegisterTime: time.Now().Unix(),
	}

	// get plugin status
	statusResp, err := requestPluginStatus(ctx, pID)
	if err != nil {
		return err
	}

	// check if it is an extension of other plugins
	for _, v := range identifyResp.MainPlugins {
		// get main plugin
		mPRoute, mEtag, err := keel.GetPluginRoute(ctx, v.ID)
		if err != nil {
			return err
		}
		if mEtag == "" {
			err = fmt.Errorf("main plugin(%s) not registered", v.ID)
			log.Error(err)
			return err
		}
		// request main plugin addons identify
		_, err = requestPluginAddonsIdentify(ctx, newPlugin.PluginID, v)
		if err != nil {
			return err
		}
		// store main plugin
		if mPRoute.Addons == nil {
			mPRoute.Addons = make(map[string]string)
		}
		for _, v := range v.Endpoints {
			mPRoute.Addons[v.AddonsPoint] = keel.EncodeRoute(pID, v.Endpoint)
		}
		err = keel.SavePluginRoute(ctx, v.ID, mPRoute, mEtag)
		if err != nil {
			log.Errorf("main plugin(%s) is busy,try again later")
			return err
		}
	}

	// get all registered plugin
	allMap, allEtag, err := keel.GetAllRegisteredPlugin(ctx)
	if err != nil {
		return err
	}
	if allEtag == "" {
		allMap = make(map[string]string)
		allEtag = "-1"
	}
	allMap[identifyResp.PluginID] = "true"
	// store new api route map and plugin and all plugins
	err = keel.SavePlugin(ctx, newPlugin, "-1")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			keel.DeletePlugin(ctx, pID)
		}
	}()

	err = keel.SavePluginRoute(ctx, pID, &keel.PluginRoute{
		Status: statusResp.Status,
	}, "1")
	if err != nil {
		return err
	}
	err = keel.SaveAllRegisteredPlugin(ctx, allMap, allEtag)
	if err != nil {
		return err
	}

	return nil
}

func parseOauth2Req(req *http.Request) (pluginID, pluginSecret string, err error) {
	contentType := req.Header.Get("content-type")
	if contentType == "application/x-www-form-urlencoded" {
		err := req.ParseForm()
		if err != nil {
			return "", "", err
		}
		pluginID = req.FormValue("client_id")
		pluginSecret = req.FormValue("client_secret")
	} else {
		auth := req.Header.Get("Authorization")
		bashAuth, err := base64.StdEncoding.DecodeString(auth)
		if err != nil {
			return "", "", err
		}
		authList := strings.Split(string(bashAuth), ":")
		if len(authList) != 2 {
			return "", "", fmt.Errorf("error invaild Authorization: %s", string(bashAuth))
		}
		pluginID = authList[0]
		pluginSecret = authList[1]
	}
	return pluginID, pluginSecret, nil
}

func checkPluginSecret(s1, s2 string) error {
	// TODO 更改校验方式
	if s1 == s2 {
		return nil
	}
	return fmt.Errorf("not match(%s/%s)", s1, s2)
}

func genPluginToken(pID string) (token, jti string, err error) {
	m := make(map[string]interface{})
	m["plugin_id"] = pID
	duration := 24 * time.Hour
	token, err = idProvider.Token("user", "", duration, &m)
	jti = m["jti"].(string)
	return
}
