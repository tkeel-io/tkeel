package config

const (
	// Port is the port of the plugin, grpc/http depending on mode.
	PluginPort string = "PLUGIN_PORT"
	// ID is the ID of the plugins.
	PluginID string = "PLUGIN_ID"
	// DaprGRPCPort is the dapr api grpc port.
	DaprGRPCPort string = "DAPR_GRPC_PORT"
	// DaprHTTPPort is the dapr api http port.
	DaprHTTPPort string = "DAPR_HTTP_PORT"
	// PluginSecretKey is random string issued by keel platform.
	PluginSecretKey string = "PLUGIN_SECRET_KEY"
)
