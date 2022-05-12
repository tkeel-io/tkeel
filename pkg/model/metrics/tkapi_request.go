package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	METRICS_LABLE_TENANT = "tenant_id"
	METRICS_LABLE_CODE   = "code"
	METRICS_LABLE_PATH   = "path"
	METRICS_LABLE_plugin = "plugin"
)

var CollectorTKApiRequest = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tkapi_request_total",
		Help: "tkeel api request counter",
	},
	[]string{METRICS_LABLE_TENANT, METRICS_LABLE_plugin, METRICS_LABLE_CODE},
)

var CollectorTKApiRequestDuration = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tkapi_request_duration",
		Help: "tkeel api request time duration",
	},
	[]string{METRICS_LABLE_TENANT},
)
