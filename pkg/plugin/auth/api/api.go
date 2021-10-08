package api

import (
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
	TenantCreate(e *openapi.APIEvent)
	TenantQuery(e *openapi.APIEvent)

	UserCreate(e *openapi.APIEvent)
	UserLogin(e *openapi.APIEvent)
	UserLogout(e *openapi.APIEvent)

	RoleCreate(e *openapi.APIEvent)
	RoleQuery(e *openapi.APIEvent)
	RolePermissionAdd(e *openapi.APIEvent)
	RolePermissionQuery(e *openapi.APIEvent)

	UserRoleAdd(e *openapi.APIEvent)
	UserRoleQuery(e *openapi.APIEvent)
	UserPermissionQuery(e *openapi.APIEvent)

	TokenCreate(e *openapi.APIEvent)
	TokenParse(e *openapi.APIEvent)
	TokenValid(e *openapi.APIEvent)
}

type api struct {
}

func (this *api) TenantQuery(e *openapi.APIEvent) {
	var (
		req      *params.TenantQueryReq
		respData *params.TenantQueryResp
	)
	req = &params.TenantQueryReq{}
	if err := utils.ReadBody2Json(e.HttpReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] UserCreate err %v", err)
		e.ResponseJson(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	tenant := model.Tenant{
		Title: req.Title,
	}
	tenants := tenant.Query(nil)
	if tenants == nil {
		log.Error("[PluginAuth] api query tenant  nil result")
		e.ResponseJson(openapi.ErrInternal)
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
		users := user.List(nil)
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
	e.ResponseJson(resp)
}

func NewApi() API {

	return &api{}
}

func (this *api) UserCreate(e *openapi.APIEvent) {
	var (
		req      *params.UserCreateReq
		respData *params.UserCreateResp
		err      error
	)
	if err = checkAuth(e.HttpReq); err != nil {
		log.Error("unauthorized access")
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrUnauthorized))
		return
	}

	req = &params.UserCreateReq{}
	if err := utils.ReadBody2Json(e.HttpReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] UserCreate err %v", err)
		e.ResponseJson(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}

	user := &model.User{
		ID:         uuid.New().String(),
		CreateTime: time.Now().Unix(),
		Name:       req.UserName,
		Password:   req.Password,
		Email:      req.Email,
	}
	err = user.Create(nil)
	if err != nil {
		log.Error("[PluginAuth] api UserCreate ", err)
		e.ResponseJson(openapi.InternalErrorResult(openapi.ErrInternal))
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
	e.ResponseJson(resp)
}

func (this *api) UserLogin(e *openapi.APIEvent) {
	var (
		req      *params.UserLoginReq
		respData *params.UserLoginResp
	)

	if e.HttpReq.Body == nil {
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrParamsInvalid))
		return
	}

	req = &params.UserLoginReq{}
	if err := utils.ReadBody2Json(e.HttpReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] UserLogin err %v", err)
		e.ResponseJson(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	if req.UserName == "" || req.Password == "" {
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrParamsInvalid))
		return
	}

	user := model.QueryUserByName(nil, req.UserName)
	if req.Password != user.Password {
		log.Error("[PluginAuth] api UserLogin password invalid")
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrParamsInvalid))
		return
	}

	token, _, err := genUserToken(user.ID, user.TenantID, "")
	if err != nil {
		log.Error(err)
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrInternal))
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
	e.ResponseJson(resp)
}

func (this *api) UserLogout(e *openapi.APIEvent) {

}

/*
租户创建  默认创建用户 以及租户管理员角色
*/
func (this *api) TenantCreate(e *openapi.APIEvent) {
	var (
		req      *params.TenantCreateReq
		respData *params.TenantCreateResp
		err      error
	)

	req = &params.TenantCreateReq{}
	if err := utils.ReadBody2Json(e.HttpReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] TenantCreate err %v", err)
		e.ResponseJson(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	//checkAuth(e.HttpReq)
	//
	//uid := e.HttpReq.Header.Get("uid")
	//if uid != "SysAdmin" {
	//	log.Warn("[PluginAuth] api TenantCre uid not SYSADMIN")
	//	e.ResponseJson(openapi.InternalErrorResult(openapi.ErrUnauthorized))
	//	return
	//}

	tenant := model.Tenant{
		Title:   req.Title,
		Email:   req.Email,
		Phone:   req.Phone,
		Country: req.Country,
		City:    req.City,
		Address: req.Address,
	}
	if err = tenant.Create(nil); err != nil {
		log.Error("[PluginAuth] api TenantCreate ", err)
		e.ResponseJson(openapi.InternalErrorResult(err.Error()))
		return
	}
	//
	user := &model.User{
		Name:     tenant.Title + "Admin",
		Password: "admin",
		TenantID: tenant.ID,
	}
	user.Create(nil)
	//
	//role := &model.Role{
	//	TenantID: tenant.ID,
	//	Name: "租户管理员",
	//	Desc:"创建租户时默认用户，赋予租户管理员权限",
	//}
	//role.Create(nil)
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
	e.ResponseJson(resp)
	return
}
func (this *api) CustomerCreate(e *openapi.APIEvent) {

}

/*

 */
func (this *api) TokenCreate(e *openapi.APIEvent) {
	var (
		req      *params.TokenCreateReq
		respData *params.TokenCreateResp
		err      error
	)

	if e.HttpReq.Body == nil {
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrParamsInvalid))
		return
	}

	checkAuth(e.HttpReq)
	req = &params.TokenCreateReq{}
	if err := utils.ReadBody2Json(e.HttpReq.Body, req); err != nil {
		log.Errorf("[api] TokenCreate err %v", err)
		e.ResponseJson(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	if req.EntityID == "" || req.EntityType == "" {
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrParamsInvalid))
		return
	}

	token, _, err := genEntityToken(e.HttpReq.Header.Get("uid"), e.HttpReq.Header.Get("tid"), "", req.EntityID, req.EntityType, nil)
	if err != nil {
		log.Error(err)
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrInternal))
		return
	}
	respData = &params.TokenCreateResp{
		token,
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		respData}

	e.ResponseJson(resp)
	return
}

/*

 */
func (this *api) TokenParse(e *openapi.APIEvent) {
	var (
		req      *params.TokenParseReq
		respData *params.TokenParseResp
		err      error
	)

	req = &params.TokenParseReq{}
	if err = utils.ReadBody2Json(e.HttpReq.Body, req); err != nil {
		log.Errorf("[api] TokenPArse err %v", err)
		e.ResponseJson(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	_, tenantID, tokenID, eid, etype, err := parseEntityToken(req.EntityToken)
	if err != nil {
		log.Error(err)
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrInternal))
		return
	}

	respData = &params.TokenParseResp{tenantID, tokenID, etype, eid}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		respData}

	e.ResponseJson(resp)
	return
}

func (this *api) TokenValid(e *openapi.APIEvent) {
	var (
		req      *params.TokenValidReq
		respData *params.TokenValidResp
	)

	req = &params.TokenValidReq{}
	if err := utils.ReadBody2Json(e.HttpReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] TokenCreate err %v", err)
		e.ResponseJson(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}

	_, _, _, _, _, err := parseEntityToken(req.EntityToken)
	if err != nil {
		log.Error(err)
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrInternal))
		return
	}

	respData = &params.TokenValidResp{true}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		respData}

	e.ResponseJson(resp)
	return

}

func (this *api) RoleCreate(e *openapi.APIEvent) {
	var (
		err error
	)
	req := &params.RoleCreateReq{}
	if err := utils.ReadBody2Json(e.HttpReq.Body, req); err != nil {
		log.Errorf("[PluginAuth] api RoleCreate ", err)
		e.ResponseJson(openapi.InternalErrorResult(openapi.ErrParamsInvalid))
		return
	}
	if err = checkAuth(e.HttpReq); err != nil {
		log.Error("[PluginAuth] api RoleCreate checkAuth ", err)
		e.ResponseJson(openapi.BadRequestResult(openapi.ErrUnauthorized))
		return
	}
	role := &model.Role{
		Name:     req.RoleName,
		Desc:     req.RoleDesc,
		TenantID: e.HttpReq.Header.Get("tid"),
	}
	if err = role.Create(nil); err != nil {
		log.Error("[PluginAuth] api RoleCreate ", err)
		e.ResponseJson(openapi.InternalErrorResult(openapi.ErrInternal))
		return
	}
	resp := &struct {
		openapi.CommonResult `json:",inline"`
		Data                 interface{} `json:"data"`
	}{openapi.SuccessResult(),
		role}
	e.ResponseJson(resp)
	return

}

func (this *api) RoleQuery(e *openapi.APIEvent) {
	panic("implement me")
}

func (this *api) RolePermissionAdd(e *openapi.APIEvent) {
	panic("implement me")
}

func (this *api) RolePermissionQuery(e *openapi.APIEvent) {
	panic("implement me")
}

func (this *api) UserRoleAdd(e *openapi.APIEvent) {
	panic("implement me")
}

func (this *api) UserRoleQuery(e *openapi.APIEvent) {
	panic("implement me")
}

func (this *api) UserPermissionQuery(e *openapi.APIEvent) {
	panic("implement me")
}
