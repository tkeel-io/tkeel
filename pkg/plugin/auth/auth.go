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
			{Endpoint: "/role/create", H: p.api.RoleCreate},
			{Endpoint: "/user/login", H: p.api.Login},
			{Endpoint: "/authenticate", H: p.api.OAuthAuthenticate},
			{Endpoint: "/user/logout", H: p.api.UserLogout},
			{Endpoint: "/user/create", H: p.api.UserCreate},
			{Endpoint: "/tenant/create", H: p.api.TenantCreate},
			{Endpoint: "/tenant/list", H: p.api.TenantQuery},
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
