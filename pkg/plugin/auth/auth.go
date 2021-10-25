package auth

import (
	"time"

	"github.com/tkeel-io/tkeel/pkg/logger"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/plugin"
	"github.com/tkeel-io/tkeel/pkg/plugin/auth/api"
	"github.com/tkeel-io/tkeel/pkg/plugin/auth/model"
)

var (
	log = logger.NewLogger("Keel.PluginAuth")
)

type PluginAuth struct {
	p   *plugin.Plugin
	api api.API
}

func NewPluginAuth(p *plugin.Plugin) *PluginAuth {
	authAPI := api.NewAPI()
	return &PluginAuth{p, authAPI}
}

func (p *PluginAuth) Run() {
	pID := p.p.Conf().Plugin.ID
	if pID == "" {
		log.Fatal("error plugin id: \"\"")
	}
	if pID != "auth" {
		log.Fatalf("error plugin id: %s should be auth", pID)
	}
	go func() {
		err := p.p.Run([]*openapi.API{
			{Endpoint: "/oauth/authenticate", H: p.api.OAuthAuthenticate},
			{Endpoint: "/oauth/token", H: p.api.OAuthToken},
			{Endpoint: "/oauth/authorize", H: p.api.OAuthAuthorize},

			{Endpoint: "/tenant/create", H: p.api.TenantCreate},
			{Endpoint: "/tenant/list", H: p.api.TenantQuery},

			{Endpoint: "/user/login", H: p.api.Login},
			{Endpoint: "/user/logout", H: p.api.UserLogout},
			{Endpoint: "/user/create", H: p.api.UserCreate},
			{Endpoint: "/user/list", H: p.api.UserRoleList},
			{Endpoint: "/user/role/add", H: p.api.UserCreate},
			{Endpoint: "/user/role/delete", H: p.api.UserCreate},
			{Endpoint: "/user/role/list", H: p.api.UserCreate},

			{Endpoint: "/role/create", H: p.api.RoleCreate},
			{Endpoint: "/role/delete", H: p.api.RoleDelete},
			{Endpoint: "/role/list", H: p.api.RoleList},
			{Endpoint: "/role/permission/add", H: p.api.RolePermissionAdd},
			{Endpoint: "/role/permission/delete", H: p.api.RolePermissionDel},
			{Endpoint: "/role/permission/list", H: p.api.RolePermissionQuery},

			{Endpoint: "/token/parse", H: p.api.TokenParse},
			{Endpoint: "/token/create", H: p.api.TokenCreate},
			{Endpoint: "/token/valid", H: p.api.TokenValid},
		}...)
		if err != nil {
			log.Fatalf("error plugin run: %s", err)
			return
		}
	}()
	time.Sleep(2 * time.Minute)
	model.UserInit()
}
