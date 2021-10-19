package plugins

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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
		openapi.APIIdentifyMethod, http.MethodGet,
		nil)
	if err != nil {
		return nil, fmt.Errorf("error call plugin: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Errorf("error response body close: %w", err)
		}
	}()

	err = utils.ReadBody2Json(resp.Body, identifyResp)
	if err != nil {
		return nil, fmt.Errorf("error read body: %w", err)
	}
	if identifyResp.Ret != 0 {
		err = fmt.Errorf("error plugin(%s) get identify: %s",
			identifyResp.PluginID, identifyResp.Msg)
		return nil, errors.New(err.Error())
	}
	// check plugin id matched.
	if identifyResp.PluginID != pID {
		err = fmt.Errorf("error plugin(%s) identify not match(%s)",
			pID, identifyResp.PluginID)
		return nil, errors.New(err.Error())
	}
	return identifyResp, nil
}

func requestPluginAddonsIdentify(ctx context.Context, pID string,
	mp *openapi.MainPlugin) (*openapi.AddonsIdentifyResp, error) {
	// check addons vaild.
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
		return nil, fmt.Errorf("error json marshal: %w", err)
	}
	addonsIdentifyResp := &openapi.AddonsIdentifyResp{}
	resp, err := keel.CallPlugin(ctx, mp.ID,
		openapi.APIAddonsIdentifyMethod, http.MethodPost,
		&keel.CallReq{
			Body: respByte,
		})
	if err != nil {
		return nil, fmt.Errorf("error json marshal: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Errorf("error response body close: %w", err)
		}
	}()

	err = utils.ReadBody2Json(resp.Body, addonsIdentifyResp)
	if err != nil {
		return nil, fmt.Errorf("error read body: %w", err)
	}
	if addonsIdentifyResp.Ret != 0 {
		return nil, fmt.Errorf("error plugin(%s) check addons identify(%s): %s",
			pID, mp.ID, addonsIdentifyResp.Msg)
	}
	return addonsIdentifyResp, nil
}

func requestPluginTenantBind(ctx context.Context, pID string,
	tenantID string, extra []byte) (*openapi.TenantBindResp, error) {
	// check addons vaild.
	tenantBindRep := &openapi.TenantBindReq{
		TenantID: tenantID,
		Extra:    extra,
	}
	reqByte, err := json.Marshal(tenantBindRep)
	if err != nil {
		return nil, fmt.Errorf("error marshal tenant bind request: %w", err)
	}
	tenantBindResp := &openapi.TenantBindResp{}
	resp, err := keel.CallPlugin(ctx, pID,
		openapi.APITenantBindMethod, http.MethodPost,
		&keel.CallReq{
			Body: reqByte,
		})
	if err != nil {
		return nil, fmt.Errorf("error post: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Errorf("error response body close: %w", err)
		}
	}()

	err = utils.ReadBody2Json(resp.Body, tenantBindResp)
	if err != nil {
		return nil, fmt.Errorf("error read body: %w", err)
	}
	if tenantBindResp.Ret != 0 {
		err = fmt.Errorf("error plugin(%s) request plugin tenant bind(%s): %s",
			pID, tenantID, tenantBindResp.Msg)
		log.Error(err)
		return nil, fmt.Errorf("error read body: %w", err)
	}
	return tenantBindResp, nil
}

func requestPluginStatus(ctx context.Context, pID string) (*openapi.StatusResp, error) {
	// TODO 在最后进行存储，而不是现在这样，插件回调就进行存储.
	statusResp := &openapi.StatusResp{}
	resp, err := keel.CallPlugin(ctx, pID, openapi.APIStatusMethod,
		http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("error call plugin: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Errorf("error response body close: %w", err)
		}
	}()

	err = utils.ReadBody2Json(resp.Body, statusResp)
	if err != nil {
		return nil, fmt.Errorf("error read body: %w", err)
	}
	if statusResp.Ret != 0 {
		return nil, fmt.Errorf("error plugin(%s) get status: %s",
			pID, statusResp.Msg)
	}
	return statusResp, nil
}

func scrapePluginStatus(ctx context.Context, interval time.Duration) error {
	// Concurrency lock.
	f, etag, err := keel.GetScrapeFlag(ctx)
	if err != nil {
		return fmt.Errorf("error get scarpe: %w", err)
	}
	if f != "" || etag != "" {
		log.Debug("failed to grab lock")
		return nil
	}
	err = keel.SaveScrapeFlag(ctx, etag, interval.Milliseconds()/1000)
	if err != nil {
		return fmt.Errorf("error save state: %w", err)
	}

	log.Debug("successfully to grab lock")
	allPluginMap, allEtag, err := keel.GetAllRegisteredPlugin(ctx)
	if err != nil {
		return fmt.Errorf("error get all registered plugin: %w", err)
	}
	for pID := range allPluginMap {
		err = scarpeStatus(ctx, pID, allPluginMap)
		if err != nil {
			log.Errorf("error scarpe status: %w", err)
			continue
		}
	}
	err = keel.SaveAllRegisteredPlugin(ctx, allPluginMap, allEtag)
	if err != nil {
		return fmt.Errorf("error save all registered plugin: %w", err)
	}
	return nil
}

func checkRegisterPluginID(ctx context.Context, identifyResp *openapi.IdentifyResp) (string, error) {
	var pID string

	if identifyResp != nil {
		pID = identifyResp.PluginID
	}
	// check if plugin has been registered.
	_, etag, err := keel.GetPlugin(ctx, pID)
	if err != nil {
		return "", fmt.Errorf("error get plugin: %w", err)
	}
	if etag != "" {
		log.Debugf("error plugin(%s) has been registered", pID)
		return "", fmt.Errorf("plugin(%s) has been registered", pID)
	}
	return pID, nil
}

func saveCachePluginRoute(ctx context.Context, pID string) error {
	pRoute, prEtag, err := keel.GetPluginRoute(ctx, pID)
	if err != nil {
		return fmt.Errorf("error get plugin route: %w", err)
	}
	if prEtag != "" {
		log.Debugf("error plugin(%s) has been registered or be registering(%s)", pID, pRoute.Status)
		return fmt.Errorf("plugin(%s) has been registered or be registering(%s)", pID, pRoute.Status)
	}

	err = keel.SavePluginRoute(ctx, pID, &keel.PluginRoute{
		Status: openapi.Starting,
	}, "-1")
	if err != nil {
		return fmt.Errorf("error save plugin route: %w", err)
	}
	return nil
}

func insertAllPlguinMap(ctx context.Context, pID string) error {
	// get all registered plugin.
	allMap, allEtag, err := keel.GetAllRegisteredPlugin(ctx)
	if err != nil {
		return fmt.Errorf("error get all resigered plugin: %w", err)
	}
	if allEtag == "" {
		allMap = make(map[string]string)
		allEtag = "-1"
	}
	allMap[pID] = "true"
	err = keel.SaveAllRegisteredPlugin(ctx, allMap, allEtag)
	if err != nil {
		return fmt.Errorf("error save all registered plugin: %w", err)
	}
	return nil
}

func deleteAllPlguinMap(ctx context.Context, pID string) error {
	// get all registered plugin.
	allMap, allEtag, err := keel.GetAllRegisteredPlugin(ctx)
	if err != nil {
		return fmt.Errorf("error get all resigered plugin: %w", err)
	}
	if allEtag == "" {
		return nil
	}
	delete(allMap, pID)
	err = keel.SaveAllRegisteredPlugin(ctx, allMap, allEtag)
	if err != nil {
		return fmt.Errorf("error save all registered plugin: %w", err)
	}
	return nil
}

// if error rollback.
func saveNewPlugin(ctx context.Context, p *keel.Plugin) error {
	// store new api route map and plugin and all plugins.
	err := keel.SavePlugin(ctx, p, "-1")
	if err != nil {
		return fmt.Errorf("error save plugin: %w", err)
	}
	defer func() {
		if err != nil {
			keel.DeletePlugin(ctx, p.PluginID)
		}
	}()

	err = insertAllPlguinMap(ctx, p.PluginID)
	if err != nil {
		return fmt.Errorf("error insert all plugin map: %w", err)
	}
	defer func() {
		if err != nil {
			deleteAllPlguinMap(ctx, p.PluginID)
		}
	}()

	// get plugin status.
	statusResp, err := requestPluginStatus(ctx, p.PluginID)
	if err != nil {
		return fmt.Errorf("error request status: %w", err)
	}
	err = keel.SavePluginRoute(ctx, p.PluginID, &keel.PluginRoute{
		Status: statusResp.Status,
	}, "1")
	if err != nil {
		return fmt.Errorf("error save plugin route: %w", err)
	}
	defer func() {
		if err != nil {
			keel.DeletePluginRoute(ctx, p.PluginID)
		}
	}()
	return nil
}

func registerPlugin(ctx context.Context, identifyResp *openapi.IdentifyResp, secret string) (err error) {
	// check plugin id vaild.
	pID, err := checkRegisterPluginID(ctx, identifyResp)
	if err != nil {
		return fmt.Errorf("error check register plugin id: %w", err)
	}

	// save cache plugin route.
	// main plugin will request new plugins.
	err = saveCachePluginRoute(ctx, pID)
	if err != nil {
		return fmt.Errorf("error save cache plugin route: %w", err)
	}
	defer func() {
		if err != nil {
			keel.DeletePluginRoute(ctx, pID)
		}
	}()

	// check if it is an extension of other plugins.
	err = registerMainPluginRoute(ctx, pID, identifyResp.MainPlugins)
	if err != nil {
		return fmt.Errorf("error register main plugin route: %w", err)
	}

	err = saveNewPlugin(ctx, &keel.Plugin{
		IdentifyResp: identifyResp,
		Secret:       secret,
		RegisterTime: time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("error save new plugin: %w", err)
	}

	return nil
}

func parseOauth2Req(req *http.Request) (pluginID, pluginSecret string, err error) {
	contentType := req.Header.Get("content-type")
	if contentType == "application/x-www-form-urlencoded" {
		err := req.ParseForm()
		if err != nil {
			return "", "", fmt.Errorf("error parse form: %w", err)
		}
		pluginID = req.FormValue("client_id")
		pluginSecret = req.FormValue("client_secret")
	} else {
		auth := req.Header.Get("Authorization")
		bashAuth, err := base64.StdEncoding.DecodeString(auth)
		if err != nil {
			return "", "", fmt.Errorf("error base64 encode: %w", err)
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
	// TODO 更改校验方式.
	if s1 == s2 {
		return nil
	}
	return fmt.Errorf("not match(%s/%s)", s1, s2)
}

func genPluginToken(pID string) (token, jti string, err error) {
	m := make(map[string]interface{})
	m["plugin_id"] = pID
	duration := 24 * time.Hour
	token, err = idProvider.Token("user", "", duration, m)
	if err != nil {
		err = fmt.Errorf("error token: %w", err)
		return
	}
	jti, ok := m["jti"].(string)
	if !ok {
		err = errors.New("error check")
		return
	}
	return
}

func registerMainPluginRoute(ctx context.Context, registerPID string, mps []*openapi.MainPlugin) error {
	for _, v := range mps {
		// get main plugin.
		mPRoute, mEtag, err := keel.GetPluginRoute(ctx, v.ID)
		if err != nil {
			return fmt.Errorf("error get plugin route: %w", err)
		}
		if mEtag == "" {
			return fmt.Errorf("main plugin(%s) not registered", v.ID)
		}
		// request main plugin addons identify.
		_, err = requestPluginAddonsIdentify(ctx, registerPID, v)
		if err != nil {
			return fmt.Errorf("error request plugin addons: %w", err)
		}
		// store main plugin.
		if mPRoute.RegisterAddons == nil {
			mPRoute.RegisterAddons = make(map[string]string)
		}
		for _, v := range v.Endpoints {
			mPRoute.RegisterAddons[v.AddonsPoint] = keel.EncodeRoute(registerPID, v.Endpoint)
		}
		err = keel.SavePluginRoute(ctx, v.ID, mPRoute, mEtag)
		if err != nil {
			log.Errorf("main plugin(%s) is busy,try again later")
			return fmt.Errorf("error save plugin route: %w", err)
		}
	}
	return nil
}

func deleteMainPluginRoute(ctx context.Context, delPluginID string, dels []*openapi.MainPlugin) error {
	for _, v := range dels {
		mpRoute, mpEtag, err := keel.GetPluginRoute(ctx, v.ID)
		if err != nil {
			return fmt.Errorf("error get plugin(%s) route: %w when delete plugin(%s) get main plugin", v.ID, err, delPluginID)
		}
		for _, ve := range v.Endpoints {
			delete(mpRoute.RegisterAddons, ve.AddonsPoint)
		}
		if mpEtag == "" {
			mpEtag = "-1"
		}
		err = keel.SavePluginRoute(ctx, v.ID, mpRoute, mpEtag)
		if err != nil {
			return fmt.Errorf("error save plugin(%s) route: %w when delete plugin(%s)", v.ID, err, delPluginID)
		}
	}
	return nil
}

func scarpeStatus(ctx context.Context, pID string, allPluginMap map[string]string) error {
	pRoute, pREtag, err := keel.GetPluginRoute(ctx, pID)
	if err != nil {
		return fmt.Errorf("error get plugin(%s) route: %w when scrape plugin status", pID, err)
	}
	statusResp, err := requestPluginStatus(ctx, pID)
	if err != nil {
		return fmt.Errorf("error request plugin(%s) status: %w when scrape plugin status", pID, err)
	}
	allPluginMap[pID] = string(statusResp.Status)
	if pRoute.Status != statusResp.Status {
		pRoute.Status = statusResp.Status
		if pREtag == "" {
			pREtag = "-1"
		}
		err := keel.SavePluginRoute(ctx, pID, pRoute, pREtag)
		if err != nil {
			return fmt.Errorf("error save plugin(%s) route: %w when scrape plugin status", pID, err)
		}
	}
	return nil
}
