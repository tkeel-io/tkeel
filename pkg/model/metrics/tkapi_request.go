package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// metrics lable.
	METRICS_LABLE_TENANT = "tenant_id"
	METRICS_LABLE_CODE   = "code"
	METRICS_LABLE_PATH   = "path"
	METRICS_LABLE_PLUGIN = "plugin"

	// metrics name.
	METRICS_NAME_TKAPI_REQUEST_TOTAL            = "tkapi_request_total"
	METRICS_NAME_TKAPI_REQUEST_DURATION_SECONDS = "tkapi_request_duration_seconds"

	METRICS_NAME_USER_TOTAL = "user_num"
	METRICS_NAME_ROLE_TOTAL = "role_num"
)

var CollectorTKApiRequest = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: METRICS_NAME_TKAPI_REQUEST_TOTAL,
		Help: "tkeel api request counter.",
	},
	[]string{METRICS_LABLE_TENANT, METRICS_LABLE_PLUGIN, METRICS_LABLE_CODE},
)

var CollectorTKApiRequestDurations = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    METRICS_NAME_TKAPI_REQUEST_DURATION_SECONDS,
		Help:    "tkapi request latency distributions.",
		Buckets: []float64{0.1, 0.2, 0.4, 0.8, 1.6, 3.2, 5.0},
	},
	[]string{METRICS_LABLE_TENANT, METRICS_LABLE_PLUGIN},
)

var CollectorUser = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: METRICS_NAME_USER_TOTAL,
		Help: "tkeel user num",
	},
	[]string{METRICS_LABLE_TENANT},
)

var CollectorRole = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: METRICS_NAME_ROLE_TOTAL,
		Help: "tkeel role num",
	},
	[]string{METRICS_LABLE_TENANT},
)
