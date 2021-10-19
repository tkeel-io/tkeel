package plugins

import (
	"context"
	"crypto/rand"
	"flag"
	"math/big"
	"time"

	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/logger"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/plugin"
	"github.com/tkeel-io/tkeel/pkg/token"
	"github.com/tkeel-io/tkeel/pkg/utils"
)

var (
	log                   = logger.NewLogger("keel.plugin.plugins")
	pluginsScrapeInterval = flag.String("plugins-scrape-interval", "30m",
		"The interval for the plugins to scrape the status of the registered plugin")
	pluginTokenSecret = flag.String("plugins-token-secret", utils.GetEnv("PLUGIN_TOKEN_SECRET", "changeme"), "gen token")
	idProvider        token.IDProvider
)

type Plugins struct {
	p *plugin.Plugin
}

func New(p *plugin.Plugin) (*Plugins, error) {
	return &Plugins{
		p: p,
	}, nil
}

func (ps *Plugins) Run() {
	pID := ps.p.Conf().Plugin.ID
	if pID == "" {
		log.Fatal("error plugin id: \"\"")
	}
	if pID != "plugins" {
		log.Fatalf("error plugin id: %s should be plugins", pID)
	}

	idProvider = token.InitIDProvider([]byte(*pluginTokenSecret), "", "")

	go func() {
		scrapeInterval, err := time.ParseDuration(*pluginsScrapeInterval)
		if err != nil {
			log.Fatalf("error parse manager-scrape-interval: %s", err)
		}
		interval := scrapeInterval
		tick := time.NewTicker(interval)
		for range tick.C {
			scrapePluginStatus(context.TODO(), scrapeInterval)
			n, err := rand.Int(rand.Reader, big.NewInt(30))
			if err != nil {
				n = big.NewInt(30)
			}
			interval = time.Duration(n.Uint64())*time.Second + scrapeInterval
			tick.Reset(interval)
		}
	}()

	go func() {
		err := ps.p.Run([]*openapi.API{
			{Endpoint: "/get", H: ps.GetPlugins},
			{Endpoint: "/list", H: ps.ListPlugins},
			{Endpoint: "/delete", H: ps.DeletePlugins},
			{Endpoint: "/register", H: ps.RegisterPlugins},
			{Endpoint: "/tenant-bind", H: ps.TenantBind},
			{Endpoint: "/oauth2/token", H: ps.Oauth2},
		}...)
		if err != nil {
			log.Fatalf("error plugin run: %s", err)
			return
		}
	}()

	log.Debugf("wait for dapr ready: %s", time.Now().Format(time.RFC3339Nano))
	if !keel.WaitDaprSidecarReady(10) {
		log.Fatalf("error dapr not ready")
	}
	log.Debugf("dapr ready: %s", time.Now().Format(time.RFC3339Nano))

	err := registerPlugin(context.TODO(), ps.p.GetIdentifyResp(), *pluginTokenSecret, ps.p.Conf().Tkeel.Version)
	if err != nil {
		log.Debugf("error register plugin plugins: %s, If its a duplicate registration error, you can ignore it", err)
	}
	log.Debugf("register plugins ok")
}
