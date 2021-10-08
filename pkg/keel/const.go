package keel

const (
	PRIVATE_STORE = "keel-private-store"
	PUBLIC_STORE  = "keel-public-store"
)

const (
	KEY_PREFIX_PLUGIN         = "plugin_"
	KEY_PREFIX_PLUGIN_ROUTE   = "plugin_route_"
	KEY_ALL_REGISTERED_PLUGIN = "all_registered_plugin"
	KEY_SCRAPE_FLAG           = "scrape_state"
)

const (
	PLUGIN_INVOKE_URL      = "http://%s/v1.0/invoke/%s/method/%s"
	ADDONS_URL_PRIFIX      = "/addons/"
	ADDONS_PATH            = "addons"
	K8S_DAPR_SIDECAR_PROBE = "http://localhost:3501/v1.0/healthz"
)
