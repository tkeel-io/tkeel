package plugin

import (
	"flag"
	"fmt"
	"strconv"

	keelconfig "github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/logger"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/utils"
	"github.com/tkeel-io/tkeel/pkg/version"
)

var (
	log = logger.NewLogger("keel.plugin")
)

type Plugin struct {
	conf *keelconfig.Configuration
	*openapi.Openapi
}

func FromFlags() (*Plugin, error) {
	pluginID := flag.String("plugin-id", utils.GetEnv("PLUGIN_ID", "keel-hello"), "Plugin id")
	pluginVersion := flag.String("plugin-version", utils.GetEnv("PLUGIN_VERSION", version.Version()), "Plugin version")
	defaultPluginHTTPPort, _ := strconv.Atoi(utils.GetEnv("PLUGIN_HTTP_PORT", "8080"))
	pluginHTTPPort := flag.Int("plugin-http-port", defaultPluginHTTPPort, "The port that the plugin listens to")
	daprPort := flag.String("dapr-http-port", utils.GetEnv("DAPR_HTTP_PORT", "3500"), "The port that the dapr listens to")
	config := flag.String("keel-plugin-config", utils.GetEnv("KEEL_PLUGIN_CONFIG", ""), "Path to config file, or name of a configuration object")

	flag.Parse()

	newPlugin := &Plugin{}

	if !keel.K8S {
		keel.SetDaprAddr("localhost:" + *daprPort)
	}

	if *config != "" {
		conf, err := keelconfig.LoadStandaloneConfiguration(*config)
		if err != nil {
			log.Errorf("load plugin config(%s) err: %s", *config, err)
			return nil, fmt.Errorf("error load plugin: %w", err)
		}
		newPlugin.conf = conf
	} else {
		conf := keelconfig.LoadDefaultConfiguration()
		conf.Plugin.ID = *pluginID
		conf.Plugin.Port = *pluginHTTPPort
		conf.Plugin.Version = *pluginVersion
		newPlugin.conf = conf
	}

	newPlugin.Openapi = openapi.NewOpenapi(newPlugin.conf.Plugin.Port, newPlugin.conf.Plugin.ID, newPlugin.conf.Plugin.Version)
	return newPlugin, nil
}

func (p *Plugin) Conf() keelconfig.Configuration {
	if p.conf != nil {
		return *p.conf
	}
	return keelconfig.Configuration{}
}

func (p *Plugin) Run(apis ...*openapi.API) error {
	for _, a := range apis {
		p.AddOpenAPI(a)
	}
	if err := p.Listen(); err != nil {
		return fmt.Errorf("error plugin listen: %w", err)
	}
	return nil
}
