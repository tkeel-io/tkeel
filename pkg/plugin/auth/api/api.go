package api

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/tkeel-io/tkeel/pkg/logger"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/plugin/auth/api/params"
	"github.com/tkeel-io/tkeel/pkg/plugin/auth/model"
	"github.com/tkeel-io/tkeel/pkg/utils"
)

var (
	log = logger.NewLogger("Keel.PluginAuth")
)

type API interface {
	OAuthToken(e *openapi.APIEvent)
	OAuthAuthorize(e *openapi.APIEvent)
	OAuthAuthenticate(e *openapi.APIEvent)
	Login(e *openapi.APIEvent)
	UserLogout(e *openapi.APIEvent)

	TenantCreate(e *openapi.APIEvent)
	TenantQuery(e *openapi.APIEvent)

	UserCreate(e *openapi.APIEvent)
	UserRoleAdd(e *openapi.APIEvent)
	UserRoleQuery(e *openapi.APIEvent)
	UserPermissionQuery(e *openapi.APIEvent)

	RoleCreate(e *openapi.APIEvent)
	RoleQuery(e *openapi.APIEvent)
	RolePermissionAdd(e *openapi.APIEvent)
	RolePermissionQuery(e *openapi.APIEvent)

	TokenCreate(e *openapi.APIEvent)
	TokenParse(e *openapi.APIEvent)
	TokenValid(e *openapi.APIEvent)
}

type api struct {
}

func NewAPI() API {
	return &api{}
}

func (a *api) OAuthToken(e *openapi.APIEvent) {
	if e.HTTPReq.Method != http.MethodGet {
		log.Errorf("error method(%s) not allowed for oauth token", e.HTTPReq.Method)
		http.Error(e, "method not allow", http.StatusMethodNotAllowed)
		return
	}
	switch utils.GetURLValue(e.HTTPReq.URL, "grant_type") {

	case "password":
		userName := utils.GetURLValue(e.HTTPReq.URL, "username")
		password := utils.GetURLValue(e.HTTPReq.URL, "password")
		if userName == "" || password == "" {
			e.ResponseJSON(openapi.BadRequestResult(openapi.ErrParamsInvalid))
			return
		}
		user := model.QueryUserByName(context.TODO(), userName)
		if password != user.Password {
			log.Error("[PluginAuth] api oauth token password invalid")
			e.ResponseJSON(openapi.BadRequestResult(openapi.ErrParamsInvalid))
			return
		}

		token, _, expire, err := genUserToken(user.ID, user.TenantID, "")
		if err != nil {
			log.Error(err)
			e.ResponseJSON(openapi.BadRequestResult(openapi.ErrInternal))
			return
		}
		respData := &params.OAuth2Token{
			AccessToken: token,
			ExpiresIn:   expire,
		}
		resp := &struct {
			openapi.CommonResult `json:",inline"`
			Data                 interface{} `json:"data"`
		}{openapi.SuccessResult(),
			respData}
		log.Info(resp)
		e.ResponseJSON(resp)
	case "code":
		fallthrough
	case "refresh_token":
		fallthrough
	default:
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrGrantTypeNotSupported))
		return
	}

}

func (a *api) OAuthAuthorize(e *openapi.APIEvent) {
	panic("implement me")
}

func (a *api) OAuthAuthenticate(e *openapi.APIEvent) {
	var (
		req      *params.UserTokenReviewReq
		respData *params.UserTokenReviewResp
		err      error
	)
	req = &params.UserTokenReviewReq{}
	if err = utils.ReadBody2Json(e.HTTPReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] api  OAuthAuthenticate err %v", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrParamsInvalid))
		return
	}
	if req.Token == "" {
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrParamsInvalid))
		return
	}
	userID, tenantID, err := parseUserToken(req.Token)
	if err != nil {
		log.Error("[PluginAuth] api OAuthAuthenticate ", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrInvalidGrant))
		return
	}
	user := model.QueryUserByID(context.TODO(), userID)
	respData = &params.UserTokenReviewResp{
		TenantID: tenantID,
		UserID:   userID,
	}
	if user != nil {
		respData.Name = user.Name
		respData.Email = user.Email
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{
		openapi.SuccessResult(),
		respData,
	}
	e.ResponseJSON(resp)
}

func (a *api) TenantQuery(e *openapi.APIEvent) {
	var (
		req      *params.TenantQueryReq
		respData *params.TenantQueryResp
	)
	req = &params.TenantQueryReq{}
	if err := utils.ReadBody2Json(e.HTTPReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] UserCreate err %v", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	tenant := model.Tenant{
		Title: req.Title,
	}
	tenants := tenant.Query(context.TODO())
	if tenants == nil {
		log.Error("[PluginAuth] api query tenant  nil result")
		e.ResponseJSON(openapi.ErrInternal)
		return
	}
	respData = &params.TenantQueryResp{}
	respData.TenantList = make([]params.TenantCreateResp, 0)
	for _, v := range tenants {
		t := params.TenantCreateResp{
			TenantID:    v.ID,
			Title:       v.Title,
			CreatedTime: v.CreatedTime,
		}
		user := model.User{
			TenantID: v.ID,
		}
		users := user.List(context.TODO())
		if users != nil {
			t.TenantAdmin = *users[0]
		}
		respData.TenantList = append(respData.TenantList, t)
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		respData}
	e.ResponseJSON(resp)
}

func (a *api) UserCreate(e *openapi.APIEvent) {
	var (
		req      *params.UserCreateReq
		respData *params.UserCreateResp
		err      error
	)
	if err = checkAuth(e.HTTPReq); err != nil {
		log.Error("unauthorized access")
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrUnauthorized))
		return
	}

	req = &params.UserCreateReq{}
	if err = utils.ReadBody2Json(e.HTTPReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] UserCreate err %v", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}

	user := &model.User{
		ID:         uuid.New().String(),
		CreateTime: time.Now().Unix(),
		Name:       req.UserName,
		Password:   req.Password,
		Email:      req.Email,
	}
	err = user.Create(context.TODO())
	if err != nil {
		log.Error("[PluginAuth] api UserCreate ", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	respData = &params.UserCreateResp{
		UserID:   user.ID,
		UserName: user.Name,
		TenantID: user.TenantID,
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		respData}
	e.ResponseJSON(resp)
}

func (a *api) Login(e *openapi.APIEvent) {
	var (
		req      *params.UserLoginReq
		respData *params.UserLoginResp
	)

	if e.HTTPReq.Body == nil {
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrParamsInvalid))
		return
	}

	req = &params.UserLoginReq{}
	if err := utils.ReadBody2Json(e.HTTPReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] Login err %v", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	if req.UserName == "" || req.Password == "" {
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrParamsInvalid))
		return
	}

	user := model.QueryUserByName(context.TODO(), req.UserName)
	if req.Password != user.Password {
		log.Error("[PluginAuth] api Login password invalid")
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrParamsInvalid))
		return
	}

	token, _, _, err := genUserToken(user.ID, user.TenantID, "")
	if err != nil {
		log.Error(err)
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrInternal))
		return
	}
	respData = &params.UserLoginResp{
		Token: token,
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		respData}
	log.Info(resp)
	e.ResponseJSON(resp)
}

func (a *api) UserLogout(e *openapi.APIEvent) {

}

/*
租户创建  默认创建用户 以及租户管理员角色.
*/
func (a *api) TenantCreate(e *openapi.APIEvent) {
	var (
		req      *params.TenantCreateReq
		respData *params.TenantCreateResp
		err      error
	)

	req = &params.TenantCreateReq{}
	if err = utils.ReadBody2Json(e.HTTPReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] TenantCreate err %v", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	// checkAuth(e.HTTPReq).

	// uid := e.HTTPReq.Header.Get("uid")
	// if uid != "SysAdmin" {
	// 	log.Warn("[PluginAuth] api TenantCre uid not SYSADMIN")
	// 	e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrUnauthorized))
	// 	return
	// }.

	tenant := model.Tenant{
		Title:   req.Title,
		Email:   req.Email,
		Phone:   req.Phone,
		Country: req.Country,
		City:    req.City,
		Address: req.Address,
	}
	if err = tenant.Create(context.TODO()); err != nil {
		log.Error("[PluginAuth] api TenantCreate ", err)
		e.ResponseJSON(openapi.InternalErrorResult(err.Error()))
		return
	}
	//
	user := &model.User{
		Name:     tenant.Title + "Admin",
		Password: "admin",
		TenantID: tenant.ID,
	}
	user.Create(context.TODO())

	// role := &model.Role{
	// 	TenantID: tenant.ID,
	// 	Name: "租户管理员",
	// 	Desc:"创建租户时默认用户，赋予租户管理员权限",
	// }
	// role.Create(nil).
	respData = &params.TenantCreateResp{
		TenantID:    tenant.ID,
		Title:       tenant.Title,
		CreatedTime: tenant.CreatedTime,
		TenantAdmin: *user,
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{
		openapi.SuccessResult(),
		respData,
	}
	e.ResponseJSON(resp)
}
func (a *api) CustomerCreate(e *openapi.APIEvent) {

}

/*
.
*/
func (a *api) TokenCreate(e *openapi.APIEvent) {
	var (
		req      *params.TokenCreateReq
		respData *params.TokenCreateResp
		err      error
	)

	if e.HTTPReq.Body == nil {
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrParamsInvalid))
		return
	}

	checkAuth(e.HTTPReq)
	req = &params.TokenCreateReq{}
	if err = utils.ReadBody2Json(e.HTTPReq.Body, req); err != nil {
		log.Errorf("[api] TokenCreate err %v", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	if req.EntityID == "" || req.EntityType == "" {
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrParamsInvalid))
		return
	}
	if req.UserID == "" {
		req.UserID = e.HTTPReq.Header.Get("uid")
	}
	if req.TenantID == "" {
		e.HTTPReq.Header.Get("tid")
	}

	token, _, err := genEntityToken(req.UserID, req.TenantID, "", req.EntityID, req.EntityType, nil)
	if err != nil {
		log.Error(err)
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrInternal))
		return
	}
	respData = &params.TokenCreateResp{
		EntityToken: token,
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		respData}

	e.ResponseJSON(resp)
}

/*
.
*/
func (a *api) TokenParse(e *openapi.APIEvent) {
	var (
		req      *params.TokenParseReq
		respData *params.TokenParseResp
		err      error
	)

	req = &params.TokenParseReq{}
	if err = utils.ReadBody2Json(e.HTTPReq.Body, req); err != nil {
		log.Errorf("[api] TokenPArse err %v", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	userID, tenantID, tokenID, eid, etype, err := parseEntityToken(req.EntityToken)
	if err != nil {
		log.Error(err)
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrInternal))
		return
	}

	respData = &params.TokenParseResp{
		UserID:     userID,
		TenantID:   tenantID,
		TokenID:    tokenID,
		EntityType: etype,
		EntityID:   eid,
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		respData}

	e.ResponseJSON(resp)
}

func (a *api) TokenValid(e *openapi.APIEvent) {
	var (
		req      *params.TokenValidReq
		respData *params.TokenValidResp
	)

	req = &params.TokenValidReq{}
	if err := utils.ReadBody2Json(e.HTTPReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] TokenCreate err %v", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}

	err := checkEntityToken(req.EntityToken)
	if err != nil {
		log.Error(err)
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrInternal))
		return
	}

	respData = &params.TokenValidResp{
		IsValid: true,
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		respData}

	e.ResponseJSON(resp)
}

func (a *api) RoleCreate(e *openapi.APIEvent) {
	var (
		err error
	)
	req := &params.RoleCreateReq{}
	if err = utils.ReadBody2Json(e.HTTPReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] api RoleCreate ", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrParamsInvalid))
		return
	}
	if err = checkAuth(e.HTTPReq); err != nil {
		log.Error("[PluginAuth] api RoleCreate checkAuth ", err)
		e.ResponseJSON(openapi.BadRequestResult(openapi.ErrUnauthorized))
		return
	}
	role := &model.Role{
		Name:     req.RoleName,
		Desc:     req.RoleDesc,
		TenantID: e.HTTPReq.Header.Get("tid"),
	}
	if err = role.Create(context.TODO()); err != nil {
		log.Error("[PluginAuth] api RoleCreate ", err)
		e.ResponseJSON(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		role}
	e.ResponseJSON(resp)
}

func (a *api) RoleQuery(e *openapi.APIEvent) {
	panic("implement me")
}

func (a *api) RolePermissionAdd(e *openapi.APIEvent) {
	panic("implement me")
}

func (a *api) RolePermissionQuery(e *openapi.APIEvent) {
	panic("implement me")
}

func (a *api) UserRoleAdd(e *openapi.APIEvent) {
	panic("implement me")
}

func (a *api) UserRoleQuery(e *openapi.APIEvent) {
	panic("implement me")
}

func (a *api) UserPermissionQuery(e *openapi.APIEvent) {
	panic("implement me")
}
