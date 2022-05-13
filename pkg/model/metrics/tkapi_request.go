package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	METRICS_LABLE_TENANT = "tenant_id"
	METRICS_LABLE_CODE   = "status_code"
	METRICS_LABLE_PATH   = "path"
	METRICS_LABLE_PLUGIN = "plugin"
)

var CollectorTKApiRequest = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tkapi_request_total",
		Help: "tkeel api request counter",
	},
	[]string{METRICS_LABLE_TENANT, METRICS_LABLE_PLUGIN, METRICS_LABLE_CODE},
)

//
//var CollectorTKApiRequestDurations = prometheus.NewCounterVec(
//	prometheus.CounterOpts{
//		Name: "tkapi_request_duration",
//		Help: "tkeel api request time duration",
//	},
//	[]string{METRICS_LABLE_TENANT, METRICS_LABLE_PLUGIN, METRICS_LABLE_CODE},
//)
//
//var CollectorTKApiRequestDurations = prometheus.NewSummaryVec(
//	prometheus.SummaryOpts{
//		Name:       "tkapi_request_durations_seconds",
//		Help:       "tkapi request latency distributions.",
//		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
//	},
//	[]string{METRICS_LABLE_TENANT, METRICS_LABLE_PLUGIN},
//)

var CollectorTKApiRequestDurations = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "rpc_durations_histogram_seconds",
		Help:    "RPC latency distributions.",
		Buckets: prometheus.LinearBuckets(0.0002-5*0.00001, .5*0.00001, 20),
	},
	[]string{METRICS_LABLE_TENANT, METRICS_LABLE_PLUGIN},
)
