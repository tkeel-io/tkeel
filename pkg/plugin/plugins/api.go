package plugins

import (
	"net/http"
	"time"

	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/utils"
)

type reqPlugin struct {
	*keel.Plugin      `json:",inline"`
	*keel.PluginRoute `json:",inline"`
}

func (ps *Plugins) GetPlugins(e *openapi.APIEvent) {
	log.Debugf("request plugins/get")

	if e.HttpReq.Method != http.MethodGet {
		log.Errorf("error method(%s) not allow", e.HttpReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	pID := utils.GetURLValue(e.HttpReq.URL, "id")

	if pID == "" {
		http.Error(e, "plugin not registered", http.StatusBadRequest)
		return
	}

	getP, _, err := keel.GetPlugin(e.HttpReq.Context(), pID)
	if err != nil {
		log.Errorf("error get plugins: %s", err)
		http.Error(e, "error get plugin", http.StatusInternalServerError)
		return
	}
	if getP == nil {
		http.Error(e, "plugin not registered", http.StatusBadRequest)
		return
	}

	pRoute, _, err := keel.GetPluginRoute(e.HttpReq.Context(), pID)
	if err != nil {
		log.Errorf("error get plugins route: %s", err)
		http.Error(e, "error get plugin route", http.StatusInternalServerError)
		return
	}

	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Datas                *reqPlugin `json:"data"`
	}{
		CommonResult: openapi.SuccessResult(),
		Datas: &reqPlugin{
			Plugin:      getP,
			PluginRoute: pRoute,
		},
	}

	ret := e.ResponseJson(resp)
	log.Debugf("get plugins: %s", string(ret))
}

func (ps *Plugins) ListPlugins(e *openapi.APIEvent) {
	log.Debugf("request plugins/list")

	if e.HttpReq.Method != http.MethodGet {
		log.Errorf("error method(%s) not allow", e.HttpReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	allMap, _, err := keel.GetAllRegisteredPlugin(e.HttpReq.Context())
	if err != nil {
		log.Errorf("error get all registered plugin: %s", err)
		http.Error(e, "error get all registered plugin", http.StatusInternalServerError)
		return
	}

	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Datas                []*reqPlugin `json:"data"`
	}{
		CommonResult: openapi.SuccessResult(),
		Datas:        make([]*reqPlugin, 0, len(allMap)),
	}

	for pID := range allMap {
		getP, _, err := keel.GetPlugin(e.HttpReq.Context(), pID)
		if err != nil {
			log.Errorf("error get plugins: %s", err)
			continue
		}
		pRoute, _, err := keel.GetPluginRoute(e.HttpReq.Context(), pID)
		if err != nil {
			log.Errorf("error get plugins route: %s", err)
			return
		}
		resp.Datas = append(resp.Datas, &reqPlugin{
			Plugin:      getP,
			PluginRoute: pRoute,
		})
	}

	ret := e.ResponseJson(resp)
	log.Debugf("get plugins: %s", string(ret))
}

func (ps *Plugins) DeletePlugins(e *openapi.APIEvent) {
	log.Debugf("request plugins/delete")

	if e.HttpReq.Method != http.MethodPost {
		log.Errorf("error method(%s) not allow", e.HttpReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	req := &struct {
		Id string `json:"id"`
	}{}
	err := utils.ReadBody2Json(e.HttpReq.Body, req)
	if err != nil {
		log.Errorf("error delete plugin: %s", err)
		http.Error(e, "error read request body", http.StatusBadRequest)
		return
	}
	defer e.HttpReq.Body.Close()

	// delete plugin and all registered plugin
	allPlugins, allEtag, err := keel.GetAllRegisteredPlugin(e.HttpReq.Context())
	if err != nil {
		log.Errorf("error delete plugin: %s", err)
		http.Error(e, "error get all registered", http.StatusInternalServerError)
		return
	}
	pID := req.Id

	delP, _, err := keel.GetPlugin(e.HttpReq.Context(), pID)
	if err != nil {
		log.Errorf("error get plugin(%s): %s", pID, err)
		http.Error(e, "error get plugin", http.StatusInternalServerError)
		return
	}
	if delP == nil {
		log.Debugf("delete plugin(%s) not registered", pID)
		e.ResponseJson(openapi.SuccessResult())
		return
	}
	for _, v := range delP.MainPlugins {
		mpRoute, mpEtag, err := keel.GetPluginRoute(e.HttpReq.Context(), v.ID)
		if err != nil {
			log.Errorf("error get plugin(%s) route: %s when delete plugin(%s) get main plugin", v.ID, err, pID)
			http.Error(e, "internal error", http.StatusInternalServerError)
			return
		}
		for _, ve := range v.Endpoints {
			delete(mpRoute.Addons, ve.AddonsPoint)
		}
		if mpEtag == "" {
			mpEtag = "-1"
		}
		err = keel.SavePluginRoute(e.HttpReq.Context(), v.ID, mpRoute, mpEtag)
		if err != nil {
			log.Errorf("error save plugin(%s) route: %s when delete plugin(%s)", v.ID, err, pID)
			http.Error(e, "internal error", http.StatusInternalServerError)
			return
		}
	}
	err = keel.DeletePlugin(e.HttpReq.Context(), pID)
	if err != nil {
		log.Errorf("error delete plugin(%s): %s", pID, err)
		http.Error(e, "error delete plugin", http.StatusInternalServerError)
		return
	}
	err = keel.DeletePluginRoute(e.HttpReq.Context(), pID)
	if err != nil {
		log.Errorf("error delete plugin(%s) route: %s", pID, err)
		http.Error(e, "error delete plugin route", http.StatusInternalServerError)
		return
	}
	delete(allPlugins, pID)

	err = keel.SaveAllRegisteredPlugin(e.HttpReq.Context(), allPlugins, allEtag)
	if err != nil {
		log.Errorf("error delete plugin: %s", err)
		http.Error(e, "error save all registered plugin", http.StatusInternalServerError)
		return
	}
	// return http succss
	e.ResponseJson(openapi.SuccessResult())
	log.Debugf("delete plugins: %v", req.Id)
}

func (ps *Plugins) RegisterPlugins(e *openapi.APIEvent) {
	log.Debugf("request plugins/register")

	if e.HttpReq.Method != http.MethodPost {
		log.Errorf("error method(%s) not allow", e.HttpReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	req := &struct {
		Id     string `json:"id"`
		Secret string `json:"secret"`
	}{}
	err := utils.ReadBody2Json(e.HttpReq.Body, req)
	if err != nil {
		log.Errorf("error register plugin: %s", err)
		http.Error(e, "error read reqeuest body", http.StatusBadRequest)
		return
	}
	defer e.HttpReq.Body.Close()

	pID := req.Id
	secret := req.Secret

	identifyResp, err := requestPluginIdentify(e.HttpReq.Context(), pID)
	if err != nil {
		log.Errorf("error request plugins identify: %s", err)
		http.Error(e, "error requst new plugin", http.StatusBadRequest)
		return
	}

	err = registerPlugin(e.HttpReq.Context(), identifyResp, secret)
	if err != nil {
		log.Errorf("error register plugins: %s", err)
		http.Error(e, "error register", http.StatusInternalServerError)
		return
	}
	// return http succss
	e.ResponseJson(openapi.SuccessResult())
	log.Debugf("register plugins: %s succ", pID)
}

func (ps *Plugins) TenantBind(e *openapi.APIEvent) {
	log.Debugf("request plugins/tenant-bind")

	if e.HttpReq.Method != http.MethodPost {
		log.Errorf("error method(%s) not allow", e.HttpReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	req := &struct {
		PluginId string `json:"plugin_id"`
		Version  string `json:"version"`
		TenantId string `json:"tenant_id"`
		License  string `json:"license"`
		Extra    []byte `json:"extra"`
	}{}
	err := utils.ReadBody2Json(e.HttpReq.Body, req)
	if err != nil {
		log.Errorf("error register plugin: %s", err)
		http.Error(e, "error read request body", http.StatusBadRequest)
		return
	}
	defer e.HttpReq.Body.Close()

	getP, getPluginEtag, err := keel.GetPlugin(e.HttpReq.Context(), req.PluginId)
	if err != nil {
		log.Errorf("error get plugins: %s", err)
		http.Error(e, "error get plugins", http.StatusBadRequest)
		return
	}
	if getP == nil {
		http.Error(e, "error plugin not registered", http.StatusBadRequest)
		return
	}

	resp, err := requestPluginTenantBind(e.HttpReq.Context(), req.PluginId, req.TenantId, req.Extra)
	if err != nil {
		log.Errorf("error request bind tenant: %s", err)
		http.Error(e, "error request bind tenant", http.StatusInternalServerError)
		return
	}
	if resp.Ret != 0 {
		log.Errorf("error request bind tenant: %s", resp.Msg)
		http.Error(e, "error reqeust plugins: "+resp.Msg, http.StatusBadRequest)
		return
	}

	if getP.ActiveTenant == nil {
		getP.ActiveTenant = make([]*keel.Tenant, 0, 1)
	}

	getP.ActiveTenant = append(getP.ActiveTenant, &keel.Tenant{
		TenantID: req.PluginId,
	})

	err = keel.SavePlugin(e.HttpReq.Context(), getP, getPluginEtag)
	if err != nil {
		log.Errorf("error save plugin: %s", err)
		http.Error(e, "error save plugin", http.StatusInternalServerError)
		return
	}

	// return http succss
	e.ResponseJson(openapi.SuccessResult())
	log.Debugf("bind tenant plugins: %s/%s", req.PluginId, req.TenantId)
}

func (ps *Plugins) Oauth2(e *openapi.APIEvent) {
	log.Debug("==start== route oauth2")
	// get oauth2 request
	pluginID, pluginSecret, err := parseOauth2Req(e.HttpReq)
	if err != nil {
		log.Errorf("error parsh oauth2 request: %s", err)
		http.Error(e, "internal error", http.StatusInternalServerError)
		return
	}

	if pluginID == "" {
		log.Debug("plugins id is empty")
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}

	var nSecret string
	if pluginID == "plugins" {
		nSecret = *pluginTokenSecret
	} else {
		// get plugin state secret
		plugin, _, err := keel.GetPlugin(e.HttpReq.Context(), pluginID)
		if err != nil {
			log.Errorf("error get plugin: %s", err)
			http.Error(e, "bad request", http.StatusBadRequest)
			return
		}
		if plugin == nil {
			log.Error("error plugin not registered")
			http.Error(e, "bad request", http.StatusBadRequest)
			return
		}
		nSecret = plugin.Secret
	}

	// check secret
	if err := checkPluginSecret(nSecret, pluginSecret); err != nil {
		log.Errorf("error plugin secret: %s", err)
		http.Error(e, "bad request", http.StatusBadRequest)
		return
	}

	// gen access token
	token, _, err := genPluginToken(pluginID)
	if err != nil {
		log.Errorf("error gen token: %s", err)
		http.Error(e, "internal error", http.StatusInternalServerError)
		return
	}

	e.ResponseJson(struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresSec   int32  `json:"expires_in"`
	}{
		AccessToken: token,
		ExpiresSec:  int32((24 * time.Hour).Seconds()),
	})
	log.Debugf("issue plugin(%s/%s) token: %s", pluginID, pluginSecret, token)
}
