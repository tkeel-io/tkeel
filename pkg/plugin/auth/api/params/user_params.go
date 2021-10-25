package params

import "github.com/tkeel-io/tkeel/pkg/plugin/auth/model"

type AuthorityType string

type UserCreateReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type UserCreateResp struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	TenantID string `json:"tenant_id"`
}

type UserLoginReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}
type UserLoginResp struct {
	Token string `json:"token"`
}
type UserTokenReviewReq struct {
	Token string `json:"token"`
}
type UserTokenReviewResp struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

type TenantCreateReq struct {
	Title   string `json:"title"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Country string `json:"country"`
	City    string `json:"city"`
	Address string `json:"address" `
}
type TenantCreateResp struct {
	TenantID    string     `json:"tenant_id"`
	Title       string     `json:"title"`
	CreatedTime int64      `json:"created_time"`
	TenantAdmin model.User `json:"tenant_admin"`
}

type TenantQueryReq struct {
	Title string `json:"title"`
}
type TenantQueryResp struct {
	TenantList []TenantCreateResp `json:"tenant_list"`
}
type CustomerCreateReq struct {
	Title   string `json:"title"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Country string `json:"country"`
	City    string `json:"city"`
	Address string `json:"address"`
}
type CustomerCreateResp struct {
	CustomerID  string `json:"customer_id"`
	Title       string `json:"title"`
	CreatedTime int64  `json:"created_time"`
}

type RoleCreateReq struct {
	TenantID string `json:"tenant_id"`
	RoleName string `json:"role_name"`
	RoleDesc string `json:"role_desc"`
}
