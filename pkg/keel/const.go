package keel

const (
	DefaultPrivateStore = "keel-private-store"
	DefaultPublicStore  = "keel-public-store"
)

const (
	KeyPrefixPlugin        = "plugin_"
	KeyPrefixPluginRoute   = "plugin_route_"
	KeyAllRegisteredPlugin = "all_registered_plugin"
	KeyScrapeFlag          = "scrape_state"
)

const (
	PluginInvokeURLFormat = "http://%s/v1.0/invoke/%s/method/%s"
	AddonsURLPrefix       = "/addons/"
	AddonsPath            = "addons"
	K8SDaprSidecarProbe   = "http://localhost:3501/v1.0/healthz"
)
